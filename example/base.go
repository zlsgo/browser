package main

import (
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/browser"
)

func main() {
	b, err := browser.New(func(o *browser.Options) {
		// o.DefaultDevice = browser.Device.Clear()
		// o.Debug = true
	})
	if err != nil {
		zlog.Error(err)
		return
	}

	zfile.Rmdir(zfile.RealPath("tmp"))

	// b.SetPageCookie([]*http.Cookie{
	// 	{
	// 		Name:    "-now",
	// 		Value:   "...." + zstring.Rand(8),
	// 		Expires: time.Now().Add(time.Hour * 24),
	// 		Domain:  ".73zls.com",
	// 	},
	// 	// {
	// 	// 	Name:    "now",
	// 	// 	Value:   "test" + zstring.Rand(8),
	// 	// 	Expires: time.Now().Add(time.Hour * 24),
	// 	// 	Domain:  "",
	// 	// },
	// })

	zlog.Dump("main")
	// b.SavePageCookie()
	err = b.Open("http://127.0.0.1:1111", func(p *browser.Page) error {
		zlog.Info(p.MustElement("title").Text())

		// zlog.Error(p.WaitLoad(time.Second * 2))
		zlog.Error(p.Screenshot("tmp/screenshot.png"))
		zlog.Error(p.ScreenshotFullPage("tmp/screenshot-full.png"))
		zlog.Error(p.MustElement("h2").Screenshot("tmp/screenshot-h2.png"))
		// zlog.Dump(p.SetCookie())

		zlog.Warn(p.ROD().HTML())
		p.SavePageCookie()

		return nil
	})
	if err != nil {
		zlog.Error(err)
	}

	zlog.Debug("dddd")
	ss, err := b.GetCookies()
	zlog.Debug(err)
	for _, v := range ss {
		zlog.Debug(v.Name, v.Value, v.Expires)
	}
	zlog.Debug("dddd")
	// zlog.Debug(b.Request("get", "http://127.0.0.1:1111/now"))
}
