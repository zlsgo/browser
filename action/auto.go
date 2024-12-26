package action

import (
	"errors"
	"reflect"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/zlsgo/browser"
)

type ActionResult struct {
	Value any            `json:"value,omitempty"`
	Name  string         `json:"name,omitempty"`
	Key   string         `json:"key,omitempty"`
	Err   string         `json:"error,omitempty"`
	Child []ActionResult `json:"child,omitempty"`
}

type ActionType interface {
	Do(p *browser.Page, parentResults ...ActionResult) (any, error)
	Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error)
}
type Actions []Action

// Run 执行 action
func (as Actions) Run(p *browser.Page, parentResults ...ActionResult) (data []ActionResult, err error) {
	data = make([]ActionResult, 0, len(as))
	keys := zarray.Map(as, func(_ int, v Action) string {
		return v.Name
	})
	if len(keys) != len(zarray.Unique(keys)) {
		return nil, errors.New("action name is not unique")
	}

	for _, action := range as {
		now := time.Now()
		res := ActionResult{Name: action.Name, Key: action.Name}
		var parent ActionResult
		if len(parentResults) > 0 {
			parent = parentResults[0]
			res.Key = parent.Key + "_" + res.Key
		} else {
			parent = res
		}

		zlog.Tips(res.Key, ": 开始执行")
		fn := func() {
			value, err := action.Action.Do(p, parent)
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

			if len(action.Next) > 0 && res.Err == "" {
				val := zreflect.ValueOf(res.Value)
				if val.Kind() == reflect.Slice {
					for i := 0; i < val.Len(); i++ {
						nres := ActionResult{
							Value: val.Index(i).Interface(),
							Name:  action.Name,
							Key:   res.Key + "_" + ztype.ToString(i+1),
						}
						child, err := action.Action.Next(p, action.Next, nres)
						if err != nil {
							res.Err = err.Error()
							break
						}
						res.Child = append(res.Child, child...)
					}
				} else {
					res.Child, err = action.Action.Next(p, action.Next, res)
					if err != nil {
						res.Err = err.Error()
					} else {
						res.Err = ""
					}
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

func (o textType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	err = zerror.TryCatch(func() error {
		s = p.MustElement(string(o)).MustText()
		return nil
	})
	return
}

func (o textType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support next action")
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

func (o InputEnterType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
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
	return nil, errors.New("not support next action")
}

type ScreenshoType struct {
	selector string
	file     string
}

// Screenshot 截图
func Screenshot(file string, selector ...string) ScreenshoType {
	s := ScreenshoType{file: file}
	if len(selector) > 0 {
		s.selector = selector[0]
	}

	return s
}

func (o ScreenshoType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	file := o.file
	if file == "" && len(parentResults) > 0 {
		file = parentResults[0].Key + ".png"
	}
	if file == "" {
		return nil, errors.New("filename is required")
	}
	file = zfile.RealPath(file)

	element, has := ExtractElement(parentResults...)
	if has {
		if o.selector != "" {
			element, has = element.Element(o.selector)
			if !has {
				return nil, errors.New("element not found")
			}
		}

		bin, err := element.ROD().Timeout(time.Second*3).Screenshot(proto.PageCaptureScreenshotFormatPng, 0)
		if err != nil {
			return nil, errors.New("screenshot failed")
		}

		return file, zfile.WriteFile(file, bin)
	}

	page, has := ExtractPage(parentResults...)
	if has {
		p = page
	}
	if o.selector != "" {
		p.MustElement(o.selector).ROD().MustScreenshot(file)
		return
	} else {
		p.ROD().MustScreenshotFullPage(file)
	}
	s = file
	return
}

func (o ScreenshoType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support next action")
}

type SleepType struct {
	timeout time.Duration
}

// Sleep 等待
func Sleep(timeout time.Duration) SleepType {
	return SleepType{timeout: timeout}
}

func (o SleepType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	time.Sleep(o.timeout)
	return
}

func (o SleepType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support next action")
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
	timeout time.Duration
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
	if a.url == "" {
		return nil, errors.New("url is required")
	}

	data = make([]ActionResult, 0, len(a.actions))
	err = a.browser.Open(a.url, func(p *browser.Page) error {
		data, err = a.actions.Run(p)
		return err
	}, func(o *browser.PageOptions) {
		if a.timeout > 0 {
			o.Timeout = a.timeout
		}
		if len(opt) > 0 {
			opt[0](o)
		}
	})

	return
}
