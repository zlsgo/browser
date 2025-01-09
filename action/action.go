package action

import (
	"context"
	"errors"
	"time"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/zlsgo/browser"
)

type TimeoutType struct {
	timeout time.Duration
}

// TimeoutAction 设置后续动作的超时时间
func TimeouAction(timeout time.Duration) TimeoutType {
	return TimeoutType{
		timeout: timeout,
	}
}

func (o TimeoutType) Do(p *browser.Page, as Actions, panicErr bool) (s any, err error) {
	*p = *p.Timeout(o.timeout)
	return nil, nil
}

type ClickType struct {
	selector string
}

// ClickAction 点击元素
func ClickAction(selector string) ClickType {
	return ClickType{
		selector: selector,
	}
}

func (o ClickType) Do(p *browser.Page, as Actions, panicErr bool) (s any, err error) {
	return nil, p.MustElement(o.selector).Click()
}

type RaceElementType struct {
	SuccessSelectors []string
	FailedSelectors  []string
	maxRetry         int
	timeout          time.Duration
}

// RaceElement 竞争元素，结果为第一个成功元素
func RaceElement(successSelectors, failedSelectors []string, maxRetry int, timeout ...time.Duration) RaceElementType {
	o := RaceElementType{
		SuccessSelectors: successSelectors,
		FailedSelectors:  failedSelectors,
		maxRetry:         maxRetry,
	}
	if len(timeout) > 0 {
		o.timeout = timeout[0]
	}
	return o
}

func (o RaceElementType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	maxRetry := o.maxRetry
	var run func() (ele *browser.Element, err error)
	run = func() (ele *browser.Element, err error) {
		page := p
		fns := make(map[string]browser.RaceElementFunc, len(o.SuccessSelectors)+len(o.FailedSelectors))
		if o.timeout > 0 {
			page = p.Timeout(o.timeout + (time.Duration(maxRetry-o.maxRetry) * time.Second))
		}
		for _, v := range append(o.SuccessSelectors, o.FailedSelectors...) {
			if v == "" {
				continue
			}
			if _, ok := fns[v]; ok {
				return nil, errors.New("selector must be unique: " + v)
			}

			fns[v] = browser.RaceElementFunc{
				Element: func(p *browser.Page) *browser.Element {
					return page.MustElement(v)
				},
				Handle: func(element *browser.Element) (retry bool, err error) {
					if zarray.Contains(o.SuccessSelectors, v) {
						return false, nil
					}
					o.maxRetry--
					if o.maxRetry > 0 {
						return true, nil
					}
					return false, errors.New("failed to find element: " + v)
				},
			}
		}

		_, ele, err = page.RaceElement(fns)
		if err == context.DeadlineExceeded {
			if o.maxRetry > 0 {
				o.maxRetry--
				err = p.Reload()
				if err != nil {
					return nil, err
				}
				return run()
			}
			return nil, err
		}
		return
	}

	return run()
}

func (o RaceElementType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support next action")
}

type ElementsType struct {
	selector string
	filter   []string
}

// Elements 获取元素, 结果为元素列表
func Elements(selector string, filter ...string) ElementsType {
	o := ElementsType{selector: selector}
	if len(filter) > 0 {
		o.filter = filter
	}
	return o
}

func (o ElementsType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	elements, has := p.Elements(o.selector, o.filter...)
	if !has {
		return nil, errors.New("not found")
	}

	return elements, nil
}

func (o ElementsType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return as.Run(p, value)
}
