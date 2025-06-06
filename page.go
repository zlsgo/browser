package browser

import (
	"context"
	_ "embed"
	"errors"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zutil"
)

type Page struct {
	ctx     context.Context
	page    *rod.Page
	browser *Browser
	Options PageOptions
	timeout time.Duration
}

func (page *Page) FromROD(p *rod.Page) *Page {
	return &Page{
		page:    p,
		browser: page.browser,
		ctx:     page.ctx,
		Options: page.Options,
	}
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

// Value 获取上下文
func (page *Page) Value(key any) any {
	return page.ctx.Value(key)
}

// WithValue 设置上下文
func (page *Page) WithValue(key any, value any) {
	page.ctx = context.WithValue(page.ctx, key, value)
}

// NavigateComplete 等待页面加载完成
func (page *Page) NavigateComplete(fn func(), d ...time.Duration) {
	wait := page.Timeout(d...).page.MustWaitNavigation()
	fn()
	wait()
	return
}

type OpenType int

const (
	OpenTypeCurrent OpenType = iota
	OpenTypeNewTab
	OpenTypeSpa
)

// WaitOpen 等待页面打开，注意手动关闭新页面
func (page *Page) WaitOpen(openType OpenType, fn func() error, d ...time.Duration) (*Page, error) {
	var wait func() (*Page, error)
	if openType == OpenTypeNewTab {
		waitNavigation := page.waitOpen(d...)
		wait = func() (*Page, error) {
			nPage, err := waitNavigation()
			if err != nil {
				return nil, err
			}

			newPage := *page
			newPage.page = nPage
			if newPage.ctx != nil {
				newPage.page = newPage.page.Context(newPage.ctx)
			}
			_, err = nPage.Activate()
			if err != nil {
				nPage.Close()
				return nil, err
			}

			return &newPage, nil
		}
	} else {
		waitNavigation := page.Timeout(d...).page.WaitNavigation(proto.PageLifecycleEventNameNetworkAlmostIdle)
		wait = func() (*Page, error) {
			waitNavigation()
			return page, nil
		}
	}

	err := fn()
	if err != nil {
		return nil, err
	}

	return wait()
}

func (p *Page) waitOpen(d ...time.Duration) func() (*rod.Page, error) {
	var (
		targetID proto.TargetTargetID
		mu       sync.Mutex
	)

	b := p.browser.Browser.Context(p.ctx)
	wait := b.Timeout(p.GetTimeout(d...)).EachEvent(func(e *proto.TargetTargetCreated) bool {
		mu.Lock()
		defer mu.Unlock()

		if targetID == "" {
			targetID = e.TargetInfo.TargetID
		}

		return e.TargetInfo.OpenerID == p.page.TargetID
	})
	return func() (*rod.Page, error) {
		wait()
		return b.PageFromTarget(targetID)
	}
}

// WaitLoad 等待页面加载
func (page *Page) WaitLoad(d ...time.Duration) (err error) {
	_, err = page.Timeout(d...).page.Eval(jsWaitLoad)
	return
}

// WaitDOMStable 等待 DOM 稳定
func (page *Page) WaitDOMStable(diff ...float64) (err error) {
	t := page.GetTimeout()
	if len(diff) > 0 {
		return page.page.Timeout(t).WaitDOMStable(time.Second, diff[0])
	}
	return page.page.Timeout(t).WaitDOMStable(time.Second, 0.01)
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

// WaitNavigation 等待页面切换
func (page *Page) WaitNavigation(fn func() error, d ...time.Duration) error {
	wait := page.Timeout(d...).page.MustWaitNavigation()
	err := fn()
	if err != nil {
		return err
	}
	wait()
	return nil
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

// Timeout 设置超时
func (page *Page) Timeout(d ...time.Duration) *Page {
	rpage := page.page
	if page.timeout != 0 {
		rpage = rpage.CancelTimeout()
	}

	var timeout time.Duration
	if len(d) > 0 {
		timeout = d[0]
	} else {
		timeout = page.GetTimeout()
	}

	if timeout != 0 && timeout >= 0 {
		rpage = rpage.Timeout(timeout)
	}

	return &Page{
		ctx:     page.ctx,
		page:    rpage,
		Options: page.Options,
		browser: page.browser,
		timeout: timeout,
	}
}

// HasElement 检查元素是否存在，不会等待元素出现
func (page *Page) HasElement(selector string) (bool, *Element) {
	has, ele, _ := page.page.Has(selector)
	if !has {
		return false, nil
	}

	return true, &Element{
		element: ele,
		page:    page,
	}
}

// Element 获取元素，会等待元素出现
func (page *Page) Element(selector string, jsRegex ...string) (ele *Element, err error) {
	var e *rod.Element

	if len(jsRegex) == 0 {
		e, err = page.page.Element(selector)
	} else {
		e, err = page.page.ElementByJS(rod.Eval(selector, jsRegex[0]))
	}
	if err != nil {
		return
	}

	return &Element{
		element: e,
		page:    page,
	}, nil
}

func (page *Page) MustElement(selector string, jsRegex ...string) (ele *Element) {
	var err error
	element, err := page.Element(selector, jsRegex...)
	if err != nil {
		panic(err)
	}

	return element
}

func (page *Page) Elements(selector string, filter ...string) (elements Elements, has bool) {
	_, err := page.Timeout().page.Element(selector)
	if err != nil {
		if errors.Is(err, &rod.ElementNotFoundError{}) {
			return Elements{}, false
		}
		return
	}

	es, _ := page.page.Elements(selector)
	has = len(es) > 0

	f := filterElements(filter...)
	for _, e := range es {
		if ok := f(e); !ok {
			continue
		}
		elements = append(elements, &Element{
			element: e,
			page:    page,
		})
	}

	return
}

func (page *Page) MustElements(selector string, filter ...string) (elements Elements) {
	element, has := page.Elements(selector, filter...)
	if !has {
		panic(&rod.ElementNotFoundError{})
	}

	return element
}

type RaceElementFunc struct {
	Element func(p *Page) (bool, *Element)
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
			var (
				ele *Element
				has bool
			)
			err := zerror.TryCatch(func() (err error) {
				has, ele = v.Element(&Page{page: p, ctx: page.ctx, Options: page.Options, browser: page.browser})
				return err
			})
			if !has {
				return nil, &rod.ElementNotFoundError{}
			}
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

	_, doErr := race.Do()
	if err == nil && doErr != nil {
		err = doErr
	}

	if err != nil {
		name = ""
		if retry {
			t := page.GetTimeout()
			err = page.Timeout(t).NavigateWaitLoad(info.URL)
			if err == nil {
				_ = page.Timeout(t).WaitDOMStable(0.1)
				return page.Timeout(t).RaceElement(elements)
			}
		}
	}

	return
}

func (page *Page) Search(query string) (ele *Element, err error) {
	sr, err := page.page.Search(query)
	if err != nil {
		return nil, err
	}
	sr.Release()

	ele = &Element{
		element: sr.First,
		page:    page,
	}

	return ele, nil
}

func (page *Page) MustSearch(query string) (ele *Element) {
	ele, err := page.Search(query)
	if err != nil {
		panic(err)
	}
	return ele
}

type PageOptions struct {
	Ctx            context.Context
	Network        func(p *proto.NetworkEmulateNetworkConditions)
	Hijack         map[string]HijackProcess
	Device         devices.Device
	Timeout        time.Duration
	MaxTime        time.Duration
	Keep           bool
	TriggerFavicon bool
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
		ctx:     page.GetContext(),
	}

	o := zutil.Optional(PageOptions{
		Timeout:        time.Second * 120,
		TriggerFavicon: true,
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

	if p.browser.options.Stealth && len(stealth) > 0 {
		p.page.MustEvalOnNewDocument(`(()=>{` + stealth + `})()`)
	}

	for i := range b.options.Scripts {
		p.page.MustEvalOnNewDocument(b.options.Scripts[i])
	}

	if err = p.NavigateLoad(url); err != nil {
		return zerror.With(err, "failed to open the page")
	}

	if process == nil {
		return nil
	}

	return zerror.TryCatch(func() error {
		if p.Options.MaxTime > 0 {
			go func() {
				timer := time.NewTimer(p.Options.MaxTime)
				defer timer.Stop()
				select {
				case <-timer.C:
					p.page.Close()
				case <-p.ctx.Done():
				case <-p.page.GetContext().Done():
				}
			}()
		}

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
