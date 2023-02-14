package main

import (
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/browser"
)

func main() {
	b, err := browser.New()
	if err != nil {
		zlog.Error(err)
		return
	}

	b.Open("https://github.com/sohaha", func(p *browser.Page) error {
		zlog.Info(p.MustElement("title").Text())
		return nil
	})

}
