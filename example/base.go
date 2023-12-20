package main

import (
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/browser"
)

func main() {
	b, err := browser.New(func(o *browser.Options) {
		o.Flags = map[string]string{"--blink-settings": "imagesEnabled=false"}
	})
	if err != nil {
		zlog.Error(err)
		return
	}

	b.Open("https://github.com/sohaha", func(p *browser.Page) error {
		zlog.Info(p.MustElement("title").Text())
		return nil
	})
}
