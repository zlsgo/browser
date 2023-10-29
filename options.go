package browser

import (
	"time"

	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/sohaha/zlsgo/zcli"
	"github.com/sohaha/zlsgo/zfile"
)

type Options struct {
	Bin           string         // Path 本地浏览器路径
	Debug         bool           // Debug 是否开启调试模式
	DebugLog      bool           // DebugLog 是否开启调试日志
	Devtools      bool           // Devtools 是否开启调试工具
	UserDataDir   string         // UserDataDir 用户数据保存目录
	UserMode      bool           // NewUserMode 是否新建用户模式
	Incognito     bool           // Incognito 是否使用隐身模式
	Timeout       time.Duration  // Timeout 超时时间
	DefaultDevice devices.Device // DefaultDevice 默认设备
	Leakless      bool           // Leakless 是否禁止 Leakless 防止报毒
	AutoKill      bool           // AutoKill 是否自动关闭浏览器
	Envs          []string       // Envs 环境变量
}

func (b *Browser) init() {
	for _, v := range []func(b *Browser){
		setBin,
		setDebug,
		setDefaultDevice,
		setUserDataDir,
		setLeakless,
		setEnv,
	} {
		v(b)
	}
}

func setEnv(b *Browser) {
	b.launcher.Env(b.options.Envs...)
}

func setLeakless(b *Browser) {
	if b.options.Leakless {
		b.launcher.Leakless(b.options.Leakless)
		return
	}
	if b.options.AutoKill && !b.options.Leakless {
		go func() {
			<-zcli.SingleKillSignal()
			b.Close()
		}()
	}
}
func setDefaultDevice(b *Browser) {
	if b.options.DefaultDevice.Title != "" {
		b.after = append(b.after, func() {
			b.Browser.DefaultDevice(b.options.DefaultDevice)
		})
	}
}

// setBin 优先使用本地浏览器
func setBin(b *Browser) {
	path := b.options.Bin
	if path == "" {
		if p, exists := launcher.LookPath(); exists {
			path = p
		}
	}
	if !zfile.FileExist(path) {
		browser := launcher.NewBrowser()
		browser.Logger = newLogger()
		bin, err := browser.Get()
		if err != nil {
			return
		}
		path = bin
	}
	b.launcher.Bin(path)
}

// setDebug 调试模式
func setDebug(b *Browser) {
	debug := b.options.Debug
	if b.options.Devtools {
		debug = true
		b.launcher.Devtools(true)
	}

	if debug {
		b.launcher.Headless(false)

		if b.options.DebugLog {
			b.after = append(b.after, func() {
				b.Browser.Trace(true)
				b.Browser.Logger(newLogger())
			})
		}
	}
}

// setUserDataDir 用户数据保存目录
func setUserDataDir(b *Browser) {
	if b.options.UserDataDir == "" {
		return
	}
	b.launcher.UserDataDir(zfile.RealPath(b.options.UserDataDir))
}
