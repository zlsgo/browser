package browser

import (
	"context"
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
	timeout time.Duration
}

// ROD 获取 rod 实例
func (page *Page) ROD() *rod.Page {
	return page.page
}

// Browser 获取浏览器实例
func (page *Page) Browser() *Browser {
	return page.browser
}

// Close 关闭页面
func (page *Page) Close() error {
	return page.page.Close()
}

// NavigateComplete 等待页面加载完成
func (page *Page) NavigateComplete(fn func(), d ...time.Duration) {
	wait := page.Timeout(d...).page.MustWaitNavigation()
	fn()
	wait()
	return
}

// WaitOpen 等待新页面打开，注意手动关闭新页面
func (page *Page) WaitOpen(fn func() error, d ...time.Duration) (*Page, error) {
	t := page.GetTimeout(d...)
	w := page.Timeout(t).page.WaitOpen()
	err := fn()
	if err != nil {
		return nil, err
	}
	nPage, err := w()
	if err != nil {
		return nil, err
	}
	newPage := *page
	newPage.page = nPage
	return &newPage, nil
}

// WaitLoad 等待页面加载
func (page *Page) WaitLoad(d ...time.Duration) (err error) {
	_, err = page.Timeout(d...).page.Eval(jsWaitLoad)
	return
}

// WaitDOMStable 等待 DOM 稳定
func (page *Page) WaitDOMStable(diff float64, d ...time.Duration) (err error) {
	t := page.GetTimeout(d...)
	if t == 0 {
		t = time.Second
	}
	return page.page.WaitDOMStable(t, diff)
}

// NavigateLoad 导航到新 url
func (page *Page) NavigateLoad(url string) (err error) {
	if url == "" {
		url = "about:blank"
	}

	if err := page.Timeout().page.Navigate(url); err != nil {
		return err
	}
	return nil
}

// NavigateWaitLoad 导航到新 url，并等待页面加载
func (page *Page) NavigateWaitLoad(url string) (err error) {
	err = page.NavigateLoad(url)

	if err == nil && url != "about:blank" {
		_, err = page.Timeout().page.Eval(jsWaitDOMContentLoad)
	}

	return err
}

// WithTimeout 包裹一个内置的超时处理
func (page *Page) WithTimeout(d time.Duration, fn func(page *Page) error) error {
	return fn(page.Timeout(d))
}

// MustWithTimeout 包裹一个内置的超时处理，如果超时会panic
func (page *Page) MustWithTimeout(d time.Duration, fn func(page *Page) error) {
	err := page.WithTimeout(d, fn)
	if err != nil {
		panic(err)
	}
}

func (page *Page) GetTimeout(d ...time.Duration) time.Duration {
	if len(d) > 0 {
		return d[0]
	} else if page.timeout != 0 {
		return page.timeout
	} else if page.Options.Timeout != 0 {
		return page.Options.Timeout
	} else if page.browser.options.Timeout != 0 {
		return page.browser.options.Timeout
	}

	return page.timeout
}

func (page *Page) Timeout(d ...time.Duration) *Page {
	p := &Page{
		page:    page.page,
		Options: page.Options,
		browser: page.browser,
		timeout: page.GetTimeout(d...),
	}

	if p.timeout != 0 {
		p.page = page.page.Timeout(p.timeout)
	}

	return p
}

func (page *Page) Element(selector string, jsRegex ...string) (ele *Element, has bool) {
	var (
		e   *rod.Element
		err error
	)

	if len(jsRegex) == 0 {
		e, err = page.page.Element(selector)
	} else {
		e = page.page.MustElementByJS(selector, jsRegex[0])
	}
	if err != nil {
		return
	}

	return &Element{
		element: e,
		page:    page,
	}, true
}

func (page *Page) MustElement(selector string, jsRegex ...string) (ele *Element) {
	var err error
	element, has := page.Element(selector, jsRegex...)
	if !has {
		err = &rod.ElementNotFoundError{}
	}
	if err != nil {
		panic(err)
	}

	return element
}

func (page *Page) Elements(selector string) (elements Elements, has bool) {
	_, err := page.Timeout().page.Element(selector)
	if err != nil {
		if errors.Is(err, &rod.ElementNotFoundError{}) {
			return Elements{}, false
		}
		return
	}

	es, _ := page.page.Elements(selector)
	has = len(es) > 0

	for _, e := range es {
		elements = append(elements, &Element{
			element: e,
			page:    page,
		})
	}

	return
}

func (page *Page) MustElements(selector string) (elements Elements) {
	element, has := page.Elements(selector)
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
				if e := zerror.TryCatch(func() error {
					retry, err = v.Handle(ele)
					return nil
				}); e != nil {
					retry = true
				}
			}
		})
	}

	if _, err := race.Do(); err != nil {
		return "", nil, err
	}

	if err == nil && retry {
		url := info.URL
		err = page.NavigateWaitLoad(url)
		if err == nil {
			return page.RaceElement(elements)
		}
	}

	return
}

type PageOptions struct {
	Ctx            context.Context
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
		if o.Ctx != nil {
			p.page = p.page.Context(o.Ctx)
		}

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

	if err = p.NavigateLoad(url); err != nil {
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

func (page *Page) Reload() error {
	return page.page.Reload()
}
