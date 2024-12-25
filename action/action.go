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

func (o RaceElementType) Do(p *browser.Page) (s any, err error) {
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
	return nil, errors.New("not support")
}

type ElementsType struct {
	selector string
}

// Elements 获取元素, 结果为元素列表
func Elements(selector string) ElementsType {
	return ElementsType{selector: selector}
}

func (o ElementsType) Do(p *browser.Page) (s any, err error) {
	// value := [][]string{}

	elements, has := p.Elements(o.selector)
	if !has {
		return nil, errors.New("not found")
	}

	return elements, nil

	// err = zerror.TryCatch(func() error {
	// 	zlog.Tips("等待页面稳定")
	// 	p.WaitDOMStable(0, 3*time.Second)
	// 	elements := p.MustElements(o.selector)
	// 	zlog.Tips("共找到", len(elements), "个结果")
	// 	for i := range elements {
	// 		v := elements[i]
	// 		index := ztype.ToString(i + 1)
	// 		err := zerror.TryCatch(func() error {
	// 			a := v.MustElement("a")
	// 			title := v.MustElement("h2").MustText()
	// 			if title == "" {
	// 				title = a.MustHTML()
	// 			}
	// 			href := a.MustProperty("href").String()
	// 			value = append(value, []string{title, href})
	// 			zlog.Tips("  点击["+index+"]", title, href)
	// 			newPage, err := p.WaitOpen(func() error {
	// 				return a.Click()
	// 			})
	// 			if err != nil {
	// 				zlog.Error("  点击["+index+"]失败", href, err.Error())
	// 				zlog.Debug(a.HTML())
	// 				return nil
	// 			}

	// 			err = zerror.TryCatch(func() error {
	// 				defer newPage.Close()
	// 				now := time.Now()
	// 				zlog.Tips("  等待页面加载完成 >>> ", newPage.GetTimeout())
	// 				err = newPage.WaitLoad()
	// 				zlog.Debug("    ", time.Since(now), err)
	// 				err = zerror.TryCatch(func() error {
	// 					newPage.ROD().MustScreenshotFullPage(index + ".png")
	// 					return nil
	// 				})
	// 				if err != nil {
	// 					err = zerror.TryCatch(func() error {
	// 						newPage.ROD().MustScrollScreenshot(index + ".png")
	// 						return nil
	// 					})
	// 					if err != nil {
	// 						zlog.Error("  截图["+index+"]失败", err.Error())
	// 					}
	// 				}
	// 				return nil
	// 			})

	// 			return nil
	// 		})
	// 		_, _ = p.ROD().Activate()
	// 		if err != nil {
	// 			zlog.Error("  执行失败", i, err.Error())
	// 		}
	// 	}
	// 	return nil
	// })

	// return value, err
}

func (o ElementsType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support")
}
