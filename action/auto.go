package action

import (
	"errors"
	"reflect"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
	"github.com/zlsgo/browser"
)

var Debug = zutil.NewBool(false)

type AutoResult []ActionResult

func (a *AutoResult) String() string {
	j, err := zjson.Marshal(a)
	if err != nil {
		return "[]"
	}
	return zstring.Bytes2String(j)
}

type ActionResult struct {
	Value any    `json:"value,omitempty"`
	Name  string `json:"name,omitempty"`
	key   string
	Err   string         `json:"error,omitempty"`
	Child []ActionResult `json:"child,omitempty"`
}

type ActionType interface {
	Do(p *browser.Page, parentResults ...ActionResult) (any, error)
	Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error)
}
type Actions []Action

// Run 执行 action
func (as Actions) Run(page *browser.Page, parentResults ...ActionResult) (data []ActionResult, err error) {
	// p := page.Timeout(page.GetTimeout())
	p := page
	as = zarray.Filter(as, func(_ int, v Action) bool {
		return v.Action != nil
	})
	data = make([]ActionResult, 0, len(as))
	keys := zarray.Map(as, func(_ int, v Action) string {
		return v.Name
	})
	if len(keys) != len(zarray.Unique(keys)) {
		return nil, errors.New("action name is not unique")
	}

	for _, action := range as {
		res := ActionResult{Name: action.Name, key: action.Name}

		var parent ActionResult
		if len(parentResults) > 0 {
			parent = parentResults[0]
			res.key = parent.key + "_" + res.key
		} else {
			parent = res
		}

		if Debug.Load() {
			zlog.Tips("执行", res.key)
		}

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
							key:   res.key + "_" + ztype.ToString(i+1),
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
		file = parentResults[0].key + ".png"
	}
	if file == "" {
		return nil, errors.New("filename is required")
	}
	file = zfile.RealPath(file)

	element, has := ExtractElement(parentResults...)
	if has {
		if o.selector != "" {
			element, err = element.Element(o.selector)
			if err != nil {
				return nil, err
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
	s = zfile.SafePath(file)
	return
}

func (o ScreenshoType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support next action")
}

type ScreenshoFullType struct {
	file string
}

var _ ActionType = ScreenshoFullType{}

// ScreenshotFullPage 截图整个页面
func ScreenshotFullPage(file string) ScreenshoFullType {
	return ScreenshoFullType{file: file}
}

func (o ScreenshoFullType) Do(p *browser.Page, parentResults ...ActionResult) (s any, err error) {
	_ = p.WaitDOMStable(0)
	file := zfile.RealPath(o.file)
	bin, err := p.ROD().Screenshot(true, nil)
	if err != nil {
		return nil, errors.New("screenshot failed")
	}

	return zfile.SafePath(file), zfile.WriteFile(file, bin)
}

func (o ScreenshoFullType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
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
	Vaidator func(p *browser.Page, value ActionResult) error
	Name     string
	Next     Actions
}

type Auto struct {
	browser *browser.Browser
	url     string
	actions Actions
	timeout time.Duration
	debug   bool
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
func (a *Auto) Start(opt ...func(o *browser.PageOptions)) (data AutoResult, err error) {
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

// Close 关闭浏览器
func (a *Auto) Close() error {
	return a.browser.Close()
}
