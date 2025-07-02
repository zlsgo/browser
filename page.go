package browser

import (
	"context"
	_ "embed"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
	"github.com/ysmood/gson"
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

// RaceElement 等待多个元素出现，返回第一个出现的元素
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

// Search 搜索元素
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

// MustSearch 搜索元素，如果出错则 panic
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

// ScrollStrategy 滚动策略类型
type ScrollStrategy int

const (
	// ScrollStrategyAuto 自动选择滚动策略
	ScrollStrategyAuto ScrollStrategy = iota
	// ScrollStrategyCenter 将元素滚动到视口中心
	ScrollStrategyCenter
	// ScrollStrategyTop 将元素滚动到视口顶部
	ScrollStrategyTop
	// ScrollStrategyVisible 仅确保元素可见（最小滚动）
	ScrollStrategyVisible
)

// NaturalScroll 自然滚动到元素位置
func (page *Page) NaturalScroll(e *Element, expectedDuration time.Duration, horizontal ...bool) error {
	// 使用自动策略，根据元素类型自动选择最合适的滚动策略
	return page.NaturalScrollWithStrategy(e, expectedDuration, ScrollStrategyAuto, horizontal...)
}

// NaturalScrollWithStrategy 使用指定策略进行自然滚动
func (page *Page) NaturalScrollWithStrategy(e *Element, expectedDuration time.Duration, strategy ScrollStrategy, horizontal ...bool) error {
	deadline := time.Now().Add(expectedDuration)
	box, err := e.Box()
	if err != nil {
		return err
	}

	var viewportInfo gson.JSON
	viewportInfo, err = page.EvalJS(`() => ({ 
		scrollX: window.scrollX, 
		scrollY: window.scrollY, 
		innerWidth: window.innerWidth, 
		innerHeight: window.innerHeight 
	})`)
	if err != nil {
		return err
	}

	currentScrollX := float64(viewportInfo.Get("scrollX").Int())
	currentScrollY := float64(viewportInfo.Get("scrollY").Int())
	viewportWidth := float64(viewportInfo.Get("innerWidth").Int())
	viewportHeight := float64(viewportInfo.Get("innerHeight").Int())

	elementCenterX := box.X + box.Width/2
	elementCenterY := box.Y + box.Height/2

	targetScrollX := elementCenterX - viewportWidth/2
	targetScrollY := elementCenterY - viewportHeight/2

	horizontalScrollDistance := targetScrollX - currentScrollX
	verticalScrollDistance := targetScrollY - currentScrollY

	isHorizontal := false
	var scrollDistance float64

	if len(horizontal) > 0 {
		isHorizontal = horizontal[0]
		if isHorizontal {
			scrollDistance = horizontalScrollDistance
		} else {
			scrollDistance = verticalScrollDistance
		}
	} else {
		elementLeft := box.X
		elementRight := box.X + box.Width
		elementTop := box.Y
		elementBottom := box.Y + box.Height

		viewportLeft := currentScrollX
		viewportRight := currentScrollX + viewportWidth
		viewportTop := currentScrollY
		viewportBottom := currentScrollY + viewportHeight

		horizontalVisible := math.Max(0, math.Min(elementRight, viewportRight)-math.Max(elementLeft, viewportLeft)) / box.Width
		verticalVisible := math.Max(0, math.Min(elementBottom, viewportBottom)-math.Max(elementTop, viewportTop)) / box.Height

		if strategy == ScrollStrategyAuto {
			tagName, err := e.TagName()
			if err == nil {
				switch tagName {
				case "img", "video", "canvas", "svg":
					if horizontalVisible < 1.0 && verticalVisible < 1.0 {
						isHorizontal = horizontalVisible < verticalVisible
					} else if horizontalVisible < 1.0 {
						isHorizontal = true
					} else if verticalVisible < 1.0 {
						isHorizontal = false
					} else {
						isHorizontal = math.Abs(horizontalScrollDistance) > math.Abs(verticalScrollDistance)
					}
				case "input", "textarea", "select", "button":
					isHorizontal = false
				case "table", "thead", "tbody", "tr", "td", "th":
					if box.Width > viewportWidth*1.2 {
						isHorizontal = true
					} else {
						// 否则优先垂直滚
						isHorizontal = false
					}
				default:
					if (horizontalVisible < verticalVisible) ||
						(math.Abs(horizontalScrollDistance) > math.Abs(verticalScrollDistance)*1.5) {
						isHorizontal = true
					} else {
						isHorizontal = false
					}
				}
			} else {
				if (horizontalVisible < verticalVisible) ||
					(math.Abs(horizontalScrollDistance) > math.Abs(verticalScrollDistance)*1.5) {
					isHorizontal = true
				} else {
					isHorizontal = false
				}
			}
		} else if strategy == ScrollStrategyCenter {
			isHorizontal = math.Abs(horizontalScrollDistance) > math.Abs(verticalScrollDistance)
		} else if strategy == ScrollStrategyTop {
			isHorizontal = false
		} else if strategy == ScrollStrategyVisible {
			isHorizontal = horizontalVisible < verticalVisible
		} else {
			if (horizontalVisible < verticalVisible) ||
				(math.Abs(horizontalScrollDistance) > math.Abs(verticalScrollDistance)*1.5) {
				isHorizontal = true
			} else {
				isHorizontal = false
			}
		}

		if isHorizontal {
			scrollDistance = horizontalScrollDistance
		} else {
			scrollDistance = verticalScrollDistance
		}
	}

	if expectedDuration == 0 || math.Abs(scrollDistance) < 100 {
		autoSteps := int(math.Max(8, math.Min(40, math.Abs(scrollDistance)/15)))

		if math.Abs(scrollDistance) > 50 {
			initialStep := scrollDistance * 0.1 * (0.8 + 0.4*rand.Float64())
			initialStepCount := int(math.Max(3, math.Min(8, math.Abs(initialStep)/10)))

			if isHorizontal {
				page.page.Mouse.Scroll(initialStep, 0, initialStepCount)
			} else {
				page.page.Mouse.Scroll(0, initialStep, initialStepCount)
			}

			time.Sleep(time.Duration(20+rand.Intn(40)) * time.Millisecond)

			mainStep := scrollDistance - initialStep
			if isHorizontal {
				page.page.Mouse.Scroll(mainStep, 0, autoSteps)
			} else {
				page.page.Mouse.Scroll(0, mainStep, autoSteps)
			}
		} else {
			if isHorizontal {
				page.page.Mouse.Scroll(scrollDistance, 0, autoSteps)
			} else {
				page.page.Mouse.Scroll(0, scrollDistance, autoSteps)
			}
		}

		return e.ROD().ScrollIntoView()
	}

	maxSteps := int(expectedDuration.Seconds() * 2)
	minSteps := int(math.Max(3, expectedDuration.Seconds()))
	distanceBasedSteps := int(math.Abs(scrollDistance) / 100)
	scrollCount := int(math.Max(float64(minSteps), math.Min(float64(maxSteps), float64(distanceBasedSteps))))

	type scrollStep struct {
		distance float64
		delay    time.Duration
	}

	steps := make([]scrollStep, 0, scrollCount)
	totalPlannedDelay := time.Duration(0)
	totalActualDistance := 0.0

	for i := 0; i < scrollCount; i++ {
		progress := float64(i) / float64(scrollCount-1)

		var speedFactor float64
		if progress < 0.5 {
			speedFactor = 4 * progress * progress
		} else {
			speedFactor = 1 - math.Pow(-2*progress+2, 2)/2
		}

		var randomFactor float64
		if progress > 0.2 && progress < 0.8 {
			randomFactor = float64(zstring.RandInt(70, 130)) / 100.0
		} else {
			randomFactor = float64(zstring.RandInt(85, 115)) / 100.0
		}
		baseStep := scrollDistance / float64(scrollCount)
		currentStep := baseStep * speedFactor * randomFactor
		totalActualDistance += currentStep

		var baseDelay time.Duration
		if progress < 0.2 || progress > 0.8 {
			baseDelay = time.Duration(70+rand.Intn(100)) * time.Millisecond
		} else {
			baseDelay = time.Duration(40+rand.Intn(80)) * time.Millisecond
		}

		pauseProb := 0.08

		if expectedDuration > time.Second*10 {
			pauseProb = 0.15
		}
		if progress > 0.3 && progress < 0.7 {
			pauseProb += 0.05
		}
		if rand.Float64() < pauseProb && i < scrollCount-1 {
			var pauseTime int
			if expectedDuration > time.Second*10 {
				pauseTime = 150 + rand.Intn(250)
			} else {
				pauseTime = 80 + rand.Intn(150)
			}

			baseDelay += time.Duration(pauseTime) * time.Millisecond
		}

		steps = append(steps, scrollStep{
			distance: currentStep,
			delay:    baseDelay,
		})

		totalPlannedDelay += baseDelay
	}

	finalAdjustment := scrollDistance - totalActualDistance
	if len(steps) > 0 {
		steps[len(steps)-1].distance += finalAdjustment
	}

	reserveRatio := 0.15 + float64(scrollCount)*0.01
	if reserveRatio > 0.3 {
		reserveRatio = 0.3
	}
	timeAdjustFactor := float64(expectedDuration) * (1 - reserveRatio) / float64(totalPlannedDelay)

	startTime := time.Now()

	for i, step := range steps {
		if time.Now().After(deadline) {
			if isHorizontal {
				currentPosInfo, _ := page.EvalJS(`() => ({ scrollX: window.scrollX })`)
				currentX := float64(currentPosInfo.Get("scrollX").Int())
				remainDist := targetScrollX - currentX

				steps := int(math.Max(5, math.Min(30, math.Abs(remainDist)/20)))
				page.page.Mouse.Scroll(remainDist, 0, steps)
			} else {
				currentPosInfo, _ := page.EvalJS(`() => ({ scrollY: window.scrollY })`)
				currentY := float64(currentPosInfo.Get("scrollY").Int())
				remainDist := targetScrollY - currentY

				steps := int(math.Max(5, math.Min(30, math.Abs(remainDist)/20)))
				page.page.Mouse.Scroll(0, remainDist, steps)
			}
			break
		}

		stepCount := int(math.Max(1, math.Abs(step.distance)/10))

		overscrollProb := 0.05

		if expectedDuration > time.Second*8 {
			overscrollProb = 0.12
		}

		progress := float64(i) / float64(len(steps)-1)
		if progress > 0.3 && progress < 0.7 {
			overscrollProb += 0.05
		}

		if math.Abs(scrollDistance) > 500 {
			overscrollProb += 0.05
		}

		if rand.Float64() < overscrollProb && i < len(steps)-2 {
			overscrollFactor := 1.3 + 0.5*rand.Float64()
			overscrollStep := step.distance * overscrollFactor
			overscrollStepCount := int(math.Max(1, math.Abs(overscrollStep)/10))

			var jitterX, jitterY float64
			if isHorizontal {
				jitterY = (rand.Float64()*2 - 1) * math.Min(10, math.Abs(overscrollStep)*0.05)
				page.page.Mouse.Scroll(overscrollStep, jitterY, overscrollStepCount)
			} else {
				jitterX = (rand.Float64()*2 - 1) * math.Min(10, math.Abs(overscrollStep)*0.05)
				page.page.Mouse.Scroll(jitterX, overscrollStep, overscrollStepCount)
			}

			time.Sleep(time.Duration(float64(step.delay) * timeAdjustFactor * (0.2 + 0.2*rand.Float64())))

			backFactor := 0.4 + 0.2*rand.Float64()
			backStep := -overscrollStep * backFactor
			backStepCount := int(math.Max(1, math.Abs(backStep)/10))

			if isHorizontal {
				jitterY = (rand.Float64()*2 - 1) * math.Min(8, math.Abs(backStep)*0.04)
				page.page.Mouse.Scroll(backStep, jitterY, backStepCount)
			} else {
				jitterX = (rand.Float64()*2 - 1) * math.Min(8, math.Abs(backStep)*0.04)
				page.page.Mouse.Scroll(jitterX, backStep, backStepCount)
			}

			time.Sleep(time.Duration(float64(step.delay) * timeAdjustFactor * (0.2 + 0.2*rand.Float64())))

			if isHorizontal {
				jitterY = (rand.Float64()*2 - 1) * math.Min(5, math.Abs(step.distance)*0.03)
				page.page.Mouse.Scroll(step.distance, jitterY, stepCount)
			} else {
				jitterX = (rand.Float64()*2 - 1) * math.Min(5, math.Abs(step.distance)*0.03)
				page.page.Mouse.Scroll(jitterX, step.distance, stepCount)
			}

			time.Sleep(time.Duration(float64(step.delay) * timeAdjustFactor * (0.3 + 0.2*rand.Float64())))
		} else {
			var jitterX, jitterY float64

			if isHorizontal {
				jitterY = (rand.Float64()*2 - 1) * math.Min(5, math.Abs(step.distance)*0.03)
				page.page.Mouse.Scroll(step.distance, jitterY, stepCount)
			} else {
				jitterX = (rand.Float64()*2 - 1) * math.Min(5, math.Abs(step.distance)*0.03)
				page.page.Mouse.Scroll(jitterX, step.distance, stepCount)
			}

			if i < len(steps)-1 {
				delayJitter := 0.9 + 0.2*rand.Float64()
				time.Sleep(time.Duration(float64(step.delay) * timeAdjustFactor * delayJitter))
			}
		}
	}

	elapsedTime := time.Since(startTime)

	remainingTime := expectedDuration - elapsedTime
	if remainingTime > 0 && time.Now().Before(deadline) {
		var remainDistCheck float64
		if isHorizontal {
			viewportCheck, _ := page.EvalJS(`() => ({ scrollX: window.scrollX })`)
			currentCheckX := float64(viewportCheck.Get("scrollX").Int())
			remainDistCheck = math.Abs(targetScrollX - currentCheckX)
		} else {
			viewportCheck, _ := page.EvalJS(`() => ({ scrollY: window.scrollY })`)
			currentCheckY := float64(viewportCheck.Get("scrollY").Int())
			remainDistCheck = math.Abs(targetScrollY - currentCheckY)
		}

		reserveTimeRatio := 0.05
		if remainDistCheck > 50 {
			reserveTimeRatio = math.Min(0.1, 0.05+remainDistCheck/1000)
		}

		waitTime := remainingTime - time.Duration(float64(expectedDuration)*reserveTimeRatio)
		deadlineWait := time.Until(deadline)
		if waitTime > deadlineWait {
			waitTime = deadlineWait
		}
		if waitTime > 0 {
			time.Sleep(waitTime)
		}
	}

	if !time.Now().After(deadline) {
		var remainingDistance float64
		if isHorizontal {
			viewportInfo, _ = page.EvalJS(`() => ({ scrollX: window.scrollX })`)
			currentX := float64(viewportInfo.Get("scrollX").Int())
			remainingDistance = targetScrollX - currentX
		} else {
			viewportInfo, _ = page.EvalJS(`() => ({ scrollY: window.scrollY })`)
			currentY := float64(viewportInfo.Get("scrollY").Int())
			remainingDistance = targetScrollY - currentY
		}

		elapsedTime = time.Since(startTime)
		remainingTime = expectedDuration - elapsedTime

		if math.Abs(remainingDistance) > 10 && time.Now().Before(deadline) {
			finalStepCount := int(math.Max(1, math.Abs(remainingDistance)/10))
			if remainingTime < 0 {
				finalStepCount = 1
			}

			if isHorizontal {
				page.page.Mouse.Scroll(remainingDistance, 0, finalStepCount)
			} else {
				page.page.Mouse.Scroll(0, remainingDistance, finalStepCount)
			}
		}
	}

	return e.ROD().ScrollIntoView()
}
