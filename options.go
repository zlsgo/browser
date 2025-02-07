package browser

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zcli"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
)

type Options struct {
	browser         *Browser
	Hijack          HijackProcess
	Flags           map[string]string
	Bin             string
	WSEndpoint      string
	UserAgent       string
	UserDataDir     string
	AcceptLanguage  string
	ProxyUrl        string
	DefaultDevice   devices.Device
	Envs            []string
	Scripts         []string
	Extensions      []string
	SlowMotion      time.Duration
	Timeout         time.Duration
	Headless        bool
	Incognito       bool
	UserMode        bool
	Devtools        bool
	IgnoreCertError bool
	autoKill        bool
	Stealth         bool
	Leakless        bool
	Debug           bool
}

func (b *Browser) init() (err error) {
	if b == nil {
		return errors.New("browser is nil")
	}

	if b.options.UserMode {
		b.launcher = launcher.NewUserMode()
		b.options.Headless = false
	} else {
		b.launcher = launcher.New()
	}

	for _, v := range []func(b *Browser){
		setBin,
		setDebug,
		setLeakless,
		setDefaultDevice,
		setUserDataDir,
		setEnv,
		setFlags,
		setExtensions,
	} {
		v(b)
	}
	b.launcher.Headless(b.options.Headless)

	if b.options.ProxyUrl != "" {
		_ = b.client.SetProxyUrl(b.options.ProxyUrl)
	}

	if b.options.UserAgent != "" || b.options.AcceptLanguage != "" {
		ua := &proto.NetworkSetUserAgentOverride{
			AcceptLanguage: "en-US,en;q=0.9",
		}
		if b.options.AcceptLanguage != "" {
			ua.AcceptLanguage = b.options.AcceptLanguage
		}
		if b.options.UserAgent != "" {
			ua.UserAgent = b.options.UserAgent
		}
		b.userAgent = ua
	}

	if b.options.WSEndpoint == "" {
		b.options.WSEndpoint, err = b.launcher.Logger(ioutil.Discard).Launch()
		if err != nil {
			if strings.Contains(err.Error(), "Failed to launch the browser") {
				errMsg := "Failed to launch the browser"
				if zutil.IsLinux() {
					if isDebian() {
						errMsg += `: sudo apt-get install --no-install-recommends -y libnss3 libxss1 libasound2t64 libxtst6 libgtk-3-0 libgbm1 ca-certificates fonts-liberation fonts-noto-color-emoji fonts-noto-cjk`
					} else {
						errMsg += ": https://pptr.dev/troubleshooting#chrome-doesnt-launch-on-linux"
					}
				}
				return errors.New(errMsg)
			}
			return err
		}
	} else {
		b.isCustomWSEndpoint = true
	}
	b.id = ztype.DecimalToAny(int(zstring.UUID()), 64)
	b.Browser = rod.New().ControlURL(b.options.WSEndpoint)

	for _, v := range b.before {
		v()
	}

	if err = b.Browser.Connect(); err != nil {
		return err
	}

	if b.options.Incognito {
		b.Browser, err = b.Browser.Incognito()
		if err != nil {
			return err
		}
	}

	if b.options.IgnoreCertError {
		_ = b.Browser.IgnoreCertErrors(true)
	}

	for _, v := range b.after {
		v()
	}

	return nil
}

func setEnv(b *Browser) {
	b.launcher.Env(b.options.Envs...)
}

func setExtensions(b *Browser) {
	extensions := strings.Join(b.options.handerExtension(), ",")
	if extensions == "" {
		return
	}

	b.launcher.Set("load-extension", extensions)
}

func setFlags(b *Browser) {
	for n, v := range b.options.Flags {
		_ = zerror.TryCatch(func() error {
			b.launcher.Set(flags.Flag(n), v)
			return nil
		})
	}
	if b.options.ProxyUrl != "" {
		b.launcher.Set(flags.ProxyServer, b.options.ProxyUrl)
	}
}

func setLeakless(b *Browser) {
	if b.id != "" {
		return
	}

	b.launcher.Leakless(b.options.Leakless)

	go func() {
		<-zcli.SingleKillSignal()

		if b.launcher.PID() != 0 {
			p, err := os.FindProcess(b.launcher.PID())
			if err == nil {
				_ = p.Kill()
			}
		}

		_ = b.Close()
		b.Cleanup()

		os.Exit(0)
	}()
}

func setDefaultDevice(b *Browser) {
	b.after = append(b.after, func() {
		if b.options.DefaultDevice.Title == "" {
			b.Browser.NoDefaultDevice()
		} else {
			b.Browser.DefaultDevice(b.options.DefaultDevice)
		}

		if v, err := b.Browser.Version(); err == nil {
			b.client.SetUserAgent(func() string {
				if b.userAgent == nil {
					return strings.Replace(v.UserAgent, "Headless", "", -1)
				}

				return b.userAgent.UserAgent
			})
		}
	})
}

// setBin 优先使用本地浏览器
func setBin(b *Browser) {
	b.launcher.Bin(getBin(b.options.Bin))
}

func getBin(path string) string {
	if path == "" {
		if p, exists := launcher.LookPath(); exists {
			path = p
		}
	}
	if !zfile.FileExist(path) {
		browser := launcher.NewBrowser()
		browser.Logger = newLogger()
		bin, err := browser.Get()
		if err == nil {
			return bin
		}

	}
	return path
}

// setDebug 调试模式
func setDebug(b *Browser) {
	debug := b.options.Debug
	if b.options.Devtools {
		debug = true
		b.launcher.Devtools(true)
	}

	if debug {
		b.after = append(b.after, func() {
			b.Browser.Trace(true)
			b.Browser.SlowMotion(b.options.SlowMotion)
			b.Browser.Logger(newLogger())
		})
	}
}

// setUserDataDir 用户数据保存目录
func setUserDataDir(b *Browser) {
	if b.options.UserMode {
		return
	}

	if b.options.UserDataDir == "" {
		b.options.UserDataDir = zfile.TmpPath() + "/browser/" + zstring.Rand(8)
	}

	b.launcher.UserDataDir(zfile.RealPath(b.options.UserDataDir))
}
