package browser

import (
	"io/ioutil"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type Browser struct {
	Browser  *rod.Browser
	launcher *launcher.Launcher
	err      error
	options  Options
	after    []func()
}

func New(opts ...func(o *Options)) (*Browser, error) {
	o := Options{
		AutoKill: true,
	}
	for _, opt := range opts {
		opt(&o)
	}

	b := &Browser{
		options: o,
	}
	if o.UserMode {
		b.launcher = launcher.NewUserMode()
	} else {
		b.launcher = launcher.New()
	}

	b.init()

	launch, err := b.launcher.Logger(ioutil.Discard).Launch()
	if err != nil {
		return nil, err
	}

	b.Browser = rod.New().ControlURL(launch)

	if err = b.Browser.Connect(); err != nil {
		return nil, err
	}

	if o.Incognito {
		b.Browser, err = b.Browser.Incognito()
		if err != nil {
			return nil, err
		}
	}

	if o.IgnoreCertError {
		b.Browser.IgnoreCertErrors(true)
	}

	for _, v := range b.after {
		v()
	}

	return b, nil
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
	return b.Browser.Close()
}

func (b *Browser) Cleanup() {
	b.launcher.Cleanup()
}
