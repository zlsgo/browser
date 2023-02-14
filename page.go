package browser

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/sohaha/zlsgo/zerror"
)

type Page struct {
	rod.Page
}

type PageOptions struct {
	Timeout time.Duration
	Device  devices.Device
	Keep    bool
	Network func(p *proto.NetworkEmulateNetworkConditions)
	Hijack  map[string]HijackProcess
}

func (b *Browser) Open(url string, process func(*Page) error, opts ...func(o *PageOptions)) error {
	if b.err != nil {
		return b.err
	}

	page, err := stealth.Page(b.Browser)
	if err != nil {
		return zerror.With(err, "新建标签页失败")
	}

	p := &Page{
		Page: *page,
	}

	o := PageOptions{
		Timeout: time.Second * 60,
	}
	{
		for _, opt := range opts {
			opt(&o)
		}
		if !o.Keep {
			defer page.Close()
		}
		if o.Timeout != 0 {
			page = page.Timeout(o.Timeout)
		} else if b.options.Timeout != 0 {
			page = page.Timeout(b.options.Timeout)
		}

		if o.Device.Title != "" {
			page = page.MustEmulate(o.Device)
		}
		if o.Network != nil {
			page.EnableDomain(proto.NetworkEnable{})
			network := proto.NetworkEmulateNetworkConditions{
				Offline:            false,
				Latency:            0,
				DownloadThroughput: -1,
				UploadThroughput:   -1,
				ConnectionType:     proto.NetworkConnectionTypeNone,
			}
			o.Network(&network)
			network.Call(page)
		}
		if len(o.Hijack) > 0 {
			stop := p.hijack(func(router *rod.HijackRouter) {
				for k, v := range o.Hijack {
					router.MustAdd(k, func(ctx *rod.Hijack) {
						ok := v(&Hijack{ctx})
						if ok {
							ctx.ContinueRequest(&proto.FetchContinueRequest{})
						}
					})
				}
			})
			defer stop()
		}
	}

	err = page.Navigate(url)
	if err != nil {
		return zerror.With(err, "打开页面失败")
	}

	err = page.WaitLoad()
	if err != nil {
		return zerror.With(err, "页面加载失败")
	}

	return zerror.TryCatch(func() error {
		return process(p)
	})
}

func (page *Page) hijack(fn func(router *rod.HijackRouter)) func() error {
	router := page.HijackRequests()
	fn(router)
	go router.Run()
	return router.Stop
}
