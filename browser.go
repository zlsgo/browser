package browser

import (
	"net/http"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zutil"
)

type Browser struct {
	err                error
	userAgent          *proto.NetworkSetUserAgentOverride
	log                *zlog.Logger
	launcher           *launcher.Launcher
	Browser            *rod.Browser
	client             *zhttp.Engine
	id                 string
	after              []func()
	before             []func()
	cookies            []*http.Cookie
	options            Options
	isCustomWSEndpoint bool
	canUserDir         bool
}

func New(opts ...func(o *Options)) (browser *Browser, err error) {
	browser = &Browser{
		client: zhttp.New(),
		log:    zlog.New(),
	}

	browser.options = zutil.Optional(Options{
		autoKill: true,
		Headless: true,
		// Stealth:  true,
		browser: browser,
		Flags: map[string]string{
			"no-sandbox":               "",
			"disable-blink-features":   "AutomationControlled",
			"no-default-browser-check": "",
			"no-first-run":             "",
			// "disable-gpu":              "",
			// "no-startup-window":        "",
			"window-position": "0,0",
		},
		IgnoreCertError: true,
	}, opts...)

	browser.Client().EnableCookie(true)

	browser.canUserDir = browser.options.UserMode || browser.options.UserDataDir != ""

	if err := browser.init(); err != nil {
		return nil, err
	}

	return browser, nil
}

func (b *Browser) Headless(enable ...bool) (bool, error) {
	headless := b.options.Headless
	if len(enable) > 0 {
		headless = enable[0]
	}

	if b.options.Headless == headless {
		return headless, nil
	}

	if !b.isCustomWSEndpoint {
		b.options.WSEndpoint = ""
	}

	b.options.Headless = headless

	if b.launcher.PID() != 0 {
		p, err := os.FindProcess(b.launcher.PID())
		if err == nil {
			_ = p.Kill()
		}
	}

	return headless, b.init()
}

func (b *Browser) Kill() {
	b.launcher.Kill()
}

func (b *Browser) NewIncognito() *Browser {
	incognito, _ := b.Browser.Incognito()
	browser := *b
	browser.Browser = incognito
	return &browser
}

func (b *Browser) Close() error {
	if b.Browser == nil {
		return nil
	}
	return b.Browser.Close()
}

func (b *Browser) Cleanup() {
	if !b.canUserDir && b.options.UserDataDir != "" {
		_ = zfile.Rmdir(b.options.UserDataDir)
	}
}

func (b *Browser) Release() {
	b.Cleanup()
}
