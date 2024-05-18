package browser

import (
	_ "embed"
	"errors"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zutil"
)

type Page struct {
	page    *rod.Page
	Options PageOptions
	browser *Browser
}

func (page *Page) ROD() *rod.Page {
	return page.page
}

func (page *Page) Browser() *Browser {
	return page.browser
}

func (page *Page) WaitLoad(d ...time.Duration) (err error) {
	_, err = page.Timeout(d...).page.Eval(jsWaitLoad)
	return
}

func (page *Page) NavigateWaitLoad(url string) (err error) {
	if url == "" {
		url = "about:blank"
	}

	if err := page.page.Navigate(url); err != nil {
		return err
	}

	if url != "about:blank" {
		_, err = page.Timeout().page.Eval(jsWaitDOMContentLoad)
	}

	return err
}

func (page *Page) Timeout(d ...time.Duration) *Page {
	p := &Page{
		page:    page.page,
		Options: page.Options,
		browser: page.browser,
	}
	if len(d) > 0 {
		p.page = page.page.Timeout(d[0])
	} else if page.Options.Timeout != 0 {
		p.page = page.page.Timeout(page.Options.Timeout)
	} else if page.browser.options.Timeout != 0 {
		p.page = page.page.Timeout(page.browser.options.Timeout)
	}

	return p
}

func (page *Page) Element(selector string, jsRegex ...string) (ele *Element, has bool, err error) {
	var e *rod.Element
	if len(jsRegex) == 0 {
		has, e, err = page.page.Has(selector)
	} else {
		has, e, err = page.page.HasR(selector, jsRegex[0])
	}
	if err != nil {
		return
	}

	ele = &Element{
		element: e,
		page:    page,
	}
	return
}

func (page *Page) MustElement(selector string, jsRegex ...string) (ele *Element) {
	element, has, err := page.Element(selector, jsRegex...)
	if !has {
		err = &rod.ElementNotFoundError{}
	}
	if err != nil {
		panic(err)
	}

	return element
}

func (page *Page) Elements(selector string) (elems Elements, has bool, err error) {
	var es rod.Elements
	es, err = page.page.Elements(selector)
	if err != nil {
		if errors.Is(err, &rod.ElementNotFoundError{}) {
			return Elements{}, false, err
		}
		return
	}

	has = len(es) > 0

	for _, e := range es {
		elems = append(elems, &Element{
			element: e,
			page:    page,
		})
	}

	return
}

func (page *Page) MustElements(selector string) (elems Elements) {
	element, has, err := page.Elements(selector)
	if err != nil {
		panic(err)
	}

	if !has {
		panic(&rod.ElementNotFoundError{})
	}

	return element
}

type RaceElementFunc struct {
	Element func(p *Page) *Element
	Handle  func(element *Element) (retry bool, err error)
}

func (page *Page) RaceElement(elements map[string]RaceElementFunc) (name string, ele *Element, err error) {
	info, ierr := page.page.Info()
	if ierr != nil {
		err = ierr
		return
	}

	race, retry := page.page.Race(), false
	for key := range elements {
		k := key
		v := elements[k]
		race = race.ElementFunc(func(p *rod.Page) (*rod.Element, error) {
			var ele *Element
			err := zerror.TryCatch(func() error {
				ele = v.Element(&Page{page: p})
				return nil
			})
			if err != nil {
				elementNotFoundError := &rod.ElementNotFoundError{}
				if err.Error() == elementNotFoundError.Error() {
					return nil, elementNotFoundError
				}
				return nil, err
			}
			return ele.element, nil

		}).MustHandle(func(element *rod.Element) {
			name = k

			ele = &Element{
				element: element,
				page:    page,
			}
			if v.Handle != nil {
				retry, err = v.Handle(ele)
			}
		})
	}

	if _, err := race.Do(); err != nil {
		return "", nil, err
	}

	if err == nil && retry {
		_ = page.WaitLoad()

		url := info.URL
		err = page.NavigateWaitLoad(url)
		if err == nil {
			_ = page.WaitLoad()
			return page.RaceElement(elements)
		}
	}

	return
}

type PageOptions struct {
	Timeout        time.Duration
	Device         devices.Device
	Keep           bool
	TriggerFavicon bool
	Network        func(p *proto.NetworkEmulateNetworkConditions)
	Hijack         map[string]HijackProcess
}

func (b *Browser) Open(url string, process func(*Page) error, opts ...func(o *PageOptions)) error {
	if b.err != nil {
		return b.err
	}

	page, err := b.Browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return zerror.With(err, "failed to create a new tab")
	}

	if b.userAgent != nil {
		_ = page.SetUserAgent(b.userAgent)
	}

	p := &Page{
		page:    page,
		browser: b,
	}

	o := zutil.Optional(PageOptions{
		Timeout: time.Second * 60,
		// Device:  devices.LaptopWithMDPIScreen,
	}, opts...)
	{
		if o.TriggerFavicon {
			_ = p.page.TriggerFavicon()
		}

		if o.Device.Title != "" {
			p.page = p.page.MustEmulate(o.Device)
		}

		if o.Network != nil {
			p.page.EnableDomain(proto.NetworkEnable{})
			network := proto.NetworkEmulateNetworkConditions{
				Offline:            false,
				Latency:            0,
				DownloadThroughput: -1,
				UploadThroughput:   -1,
				ConnectionType:     proto.NetworkConnectionTypeNone,
			}
			o.Network(&network)
			_ = network.Call(p.page)
		}

		if b.options.Hijack != nil || len(o.Hijack) > 0 {
			stop := p.hijack(func(router *rod.HijackRouter) {
				for k, v := range o.Hijack {
					_ = router.Add(k, "", func(ctx *rod.Hijack) {
						hijaclProcess(newHijacl(ctx, b.client), v)
					})
				}

				if b.options.Hijack != nil {
					_ = router.Add("*", "", func(ctx *rod.Hijack) {
						hijaclProcess(newHijacl(ctx, b.client), b.options.Hijack)
					})
				}

				_ = router.Add("*", "", func(ctx *rod.Hijack) {
					hijaclProcess(newHijacl(ctx, b.client), func(router *Hijack) (stop bool) {
						return false

					})
				})

			})
			if !o.Keep {
				defer func() {
					_ = stop()
				}()
			}
		}
	}

	defer func() {
		if !o.Keep || err != nil {
			_ = p.page.Close()
		}
	}()

	p.Options = o

	if p.browser.options.Stealth {
		p.page.MustEvalOnNewDocument(`(()=>{` + stealth + `})()`)
	}

	for i := range b.options.Scripts {
		p.page.MustEvalOnNewDocument(b.options.Scripts[i])
	}

	if err = p.NavigateWaitLoad(url); err != nil {
		return zerror.With(err, "failed to open the page")
	}

	if b.userAgent == nil {
		_ = b.setUserAgent(p)
	}

	if process == nil {
		return nil
	}

	return zerror.TryCatch(func() error {
		return process(p)
	})
}

func hijaclProcess(h *Hijack, p HijackProcess) {
	if h.CustomState != nil {
		return
	}

	h.Skip = false

	stop := p(h)

	if h.abort {
		h.CustomState = true
		h.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return
	}

	if h.Skip {
		return
	}

	if !stop {
		h.ContinueRequest(&proto.FetchContinueRequest{})
	} else {
		h.Skip = true
	}

	h.CustomState = true
}

func (page *Page) hijack(fn func(router *rod.HijackRouter)) func() error {
	router := page.page.HijackRequests()
	fn(router)
	go router.Run()
	return router.Stop
}
