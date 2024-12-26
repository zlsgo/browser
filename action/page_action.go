package action

import (
	"errors"
	"time"

	"github.com/zlsgo/browser"
)

type waitDOMStableType struct {
	timeout time.Duration
	diff    float64
}

var _ ActionType = waitDOMStableType{}

// WaitDOMStable 等待页面稳定
func WaitDOMStable(diff float64, d ...time.Duration) waitDOMStableType {
	o := waitDOMStableType{
		diff: diff,
	}
	if len(d) > 0 {
		o.timeout = d[0]
	}
	return o
}

func (o waitDOMStableType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	if o.timeout > 0 {
		return nil, p.WaitDOMStable(o.diff, o.timeout)
	}
	return nil, p.WaitDOMStable(o.diff)
}

func (o waitDOMStableType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support next action")
}

type ClickNewPageType struct {
	selector string
}

var _ ActionType = ClickNewPageType{}

// ClickNewPage 点击新页面
func ClickNewPage(selector string) ClickNewPageType {
	return ClickNewPageType{
		selector: selector,
	}
}

func (o ClickNewPageType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	element, has := ExtractElement(parentResults...)
	if !has {
		element, has = p.Element(o.selector)
	} else if o.selector != "" {
		element, has = element.Element(o.selector)
	}

	if !has {
		return nil, errors.New("not found")
	}

	page, err := p.WaitOpen(func() error {
		return element.Click()
	})
	return page, err
}

func (o ClickNewPageType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return as.Run(p, value)
}

type ClosePageType struct{}

var _ ActionType = ClosePageType{}

// ClosePage 关闭页面
func ClosePage() ClosePageType {
	return ClosePageType{}
}

func (o ClosePageType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	if len(parentResults) == 0 {
		return nil, p.Close()
	}
	page, has := parentResults[0].Value.(*browser.Page)
	if !has {
		return nil, errors.New("not found")
	}
	page.Close()
	return nil, nil
}

func (o ClosePageType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support next action")
}

type ActivatePageType struct{}

var _ ActionType = ActivatePageType{}

// ActivatePage 激活页面
func ActivatePage() ActivatePageType {
	return ActivatePageType{}
}

func (o ActivatePageType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	page, has := ExtractPage(parentResults...)
	if !has {
		_, err := p.ROD().Activate()
		return nil, err
	}

	_, err = page.ROD().Activate()
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (o ActivatePageType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return as.Run(p, value)
}
