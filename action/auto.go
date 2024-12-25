package action

import (
	"errors"
	"time"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/browser"
)

type ActionResult struct {
	Value any            `json:"value,omitempty"`
	Name  string         `json:"name,omitempty"`
	Err   string         `json:"error,omitempty"`
	Child []ActionResult `json:"child,omitempty"`
}

type ActionType interface {
	Do(p *browser.Page) (any, error)
	Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error)
}
type Actions []Action

// Run 执行 action
func (as Actions) Run(p *browser.Page) (data []ActionResult, err error) {
	data = make([]ActionResult, 0, len(as))
	keys := zarray.Map(as, func(_ int, v Action) string {
		return v.Name
	})
	if len(keys) != len(zarray.Unique(keys)) {
		return nil, errors.New("action name is not unique")
	}

	for _, action := range as {
		zlog.Tips(action.Name, ": 开始执行")
		now := time.Now()
		res := ActionResult{Name: action.Name}
		fn := func() {
			value, err := action.Action.Do(p)
			res.Value = value
			if err != nil {
				res.Err = err.Error()
			}
			if action.Vaidator != nil {
				err = action.Vaidator(p, res)
				if err != nil {
					res.Err = err.Error()
				} else {
					res.Err = ""
				}
			}

			if len(action.Next) > 0 {
				res.Child, err = action.Action.Next(p, action.Next, res)
				if err != nil {
					res.Err = err.Error()
				} else {
					res.Err = ""
				}
			}

			data = append(data, res)
		}
		fn()
		zlog.Tips(action.Name, "执行结束: ", time.Since(now))
		if res.Err != "" {
			break
		}
	}
	return
}

type textType string

var _ ActionType = textType("")

// TextAction 获取元素文本
func TextAction(selector string) textType {
	return textType(selector)
}

func (o textType) Do(p *browser.Page) (s any, err error) {
	err = zerror.TryCatch(func() error {
		s = p.MustElement(string(o)).MustText()
		return nil
	})
	return
}

func (o textType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support")
}

type InputEnterType struct {
	text     string
	selector string
}

// InputEnter 输入文本
func InputEnter(selector, text string) InputEnterType {
	return InputEnterType{
		text:     text,
		selector: selector,
	}
}

func (o InputEnterType) Do(p *browser.Page) (s any, err error) {
	err = zerror.TryCatch(func() error {
		e, has := p.MustElement("body").Timeout().FindTextInputElement(o.selector)
		if !has {
			return errors.New("input not found")
		}
		e.InputText(o.text, true)
		return e.InputEnter()
	})
	s = o.text
	return
}

func (o InputEnterType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support")
}

type ScreenshoType struct {
	selector string
	file     string
}

// Screenshot 截图
func Screenshot(file string, selector ...string) ScreenshoType {
	s := ScreenshoType{file: zfile.RealPath(file)}
	if len(selector) > 0 {
		s.selector = selector[0]
	}

	return s
}

func (o ScreenshoType) Do(p *browser.Page) (s any, err error) {
	if o.selector != "" {
		p.MustElement(o.selector).ROD().MustScreenshot(o.file)
		return
	} else {
		p.ROD().MustScreenshotFullPage(o.file)
	}
	s = o.file
	return
}

func (o ScreenshoType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support")
}

type SleepType struct {
	timeout time.Duration
}

// Sleep 等待
func Sleep(timeout time.Duration) SleepType {
	return SleepType{timeout: timeout}
}

func (o SleepType) Do(p *browser.Page) (s any, err error) {
	time.Sleep(o.timeout)
	return
}

func (o SleepType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support")
}

type Action struct {
	Action   ActionType
	Name     string
	Next     Actions
	Vaidator func(p *browser.Page, value ActionResult) error
}

type Auto struct {
	browser *browser.Browser
	actions Actions
	url     string
}

// NewAuto 创建自动执行器
func NewAuto(b *browser.Browser, url string, actions []Action) *Auto {
	return &Auto{
		browser: b,
		url:     url,
		actions: actions,
	}
}

// Start 开始执行
func (a *Auto) Start(opt ...func(o *browser.PageOptions)) (data []ActionResult, err error) {
	data = make([]ActionResult, 0, len(a.actions))
	err = a.browser.Open(a.url, func(p *browser.Page) error {
		data, err = a.actions.Run(p)
		return err
	}, func(o *browser.PageOptions) {
		if len(opt) > 0 {
			opt[0](o)
		}
	})

	return
}
