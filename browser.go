package browser

import (
	"errors"
	"net/http"
	"os"
	"strings"

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
			"disable-component-update": "",
			"window-position":          "0,0",
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

// SetCookie set global cookies
func (b *Browser) SetCookies(cookies []*http.Cookie) error {
	if cookies == nil {
		b.cookies = make([]*http.Cookie, 0, 0)
		_ = b.Browser.SetCookies(nil)
		return nil
	}

	b.cookies = b.uniqueCookies(cookies)
	c, err := b.cookiesToProto(cookies)
	if err != nil {
		return errors.New("failed to set cookie: " + err.Error())
	}

	b.Browser.SetCookies(c)
	return nil
}

// GetCookie get global cookies
func (b *Browser) GetCookies() ([]*http.Cookie, error) {
	protoCookies, err := b.Browser.GetCookies()
	if err != nil {
		return []*http.Cookie{}, err
	}

	cookies := make([]*http.Cookie, 0, len(protoCookies))
	for i := range protoCookies {
		value := protoCookies[i].Value
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}
		cookie := http.Cookie{
			Name:     protoCookies[i].Name,
			Value:    value,
			Path:     protoCookies[i].Path,
			Domain:   protoCookies[i].Domain,
			Secure:   protoCookies[i].Secure,
			HttpOnly: protoCookies[i].HTTPOnly,
		}
		if protoCookies[i].Expires > 0 {
			cookie.Expires = protoCookies[i].Expires.Time()
		}
		cookies = append(cookies, &cookie)
	}

	return cookies, nil
}

func (b *Browser) cookiesToProto(cookies []*http.Cookie) ([]*proto.NetworkCookieParam, error) {
	protoCookies := make([]*proto.NetworkCookieParam, 0, len(cookies))
	for i := range cookies {
		if cookies[i].Domain == "" {
			return nil, errors.New("domain is required for cookie configuration")
		}
		if cookies[i].Name == "" {
			return nil, errors.New("name is required for cookie configuration")
		}

		protoCookies = append(protoCookies, &proto.NetworkCookieParam{
			Name:     cookies[i].Name,
			Value:    cookies[i].Value,
			Expires:  proto.TimeSinceEpoch(cookies[i].Expires.Unix()),
			Path:     cookies[i].Path,
			Domain:   cookies[i].Domain,
			Secure:   cookies[i].Secure,
			HTTPOnly: cookies[i].HttpOnly,
		})
	}

	return protoCookies, nil
}
