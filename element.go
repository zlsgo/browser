package browser

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/js"
)

type Element struct {
	element *rod.Element
	page    *Page
}

type Elements []*Element

func (p *Page) Document() (*Element, error) {
	jsElement := &js.Function{
		Name:       "element",
		Definition: `function(e){return document.body}`,
	}
	e, err := p.ROD().ElementByJS(&rod.EvalOptions{
		ByValue: true,
		JSArgs:  []interface{}{jsElement},
		JS:      fmt.Sprintf(`function (f /* %s */, ...args) { return f.apply(this, args) }`, jsElement.Name),
	})
	if err != nil {
		return nil, err
	}

	return &Element{
		element: e,
		page:    p,
	}, nil
}

func (e *Element) ROD() *rod.Element {
	return e.element
}

func (e *Element) Timeout(d ...time.Duration) *Element {
	element := e.element
	if e.page.timeout != 0 {
		element = element.CancelTimeout()
	}

	if len(d) > 0 {
		if d[0] >= 0 {
			element = element.Timeout(d[0])
		}
	} else if e.page.timeout != 0 {
		if e.page.timeout >= 0 {
			element = element.Timeout(e.page.timeout)
		}
	}

	return &Element{
		element: element,
		page:    e.page,
	}
}

// HasElement 检查元素是否存在，不会等待元素出现
func (e *Element) Parent() (element *Element, err error) {
	ele, err := e.element.Parent()
	if err != nil {
		return nil, err
	}

	return &Element{
		element: ele,
		page:    e.page,
	}, nil
}

// Frame 获取元素的 iframe 页面
func (e *Element) Frame() (*Page, error) {
	frame, err := e.element.Frame()
	if err != nil {
		return nil, err
	}

	return e.page.FromROD(frame), nil
}

// HasElement 检查元素是否存在，不会等待元素出现
func (e *Element) HasElement(selector string) (bool, *Element) {
	has, ele, _ := e.element.Has(selector)
	if !has {
		return false, nil
	}

	return true, &Element{
		element: ele,
		page:    e.page,
	}
}

// Element 获取元素，会等待元素出现
func (e *Element) Element(selector string, jsRegex ...string) (element *Element, err error) {
	var (
		relm *rod.Element
	)
	if len(jsRegex) == 0 {
		relm, err = e.element.Element(selector)
	} else {
		relm, err = e.element.ElementR(selector, jsRegex[0])
	}

	if err != nil {
		return nil, err
	}

	return &Element{
		element: relm,
		page:    e.page,
	}, nil
}

func (e *Element) MustElement(selector string, jsRegex ...string) *Element {
	elm, err := e.Element(selector, jsRegex...)
	if err != nil {
		panic(err)
	}
	return elm
}

func (e *Element) Elements(selector string) (elements Elements, has bool) {
	_, err := e.element.Element(selector)
	if err != nil {
		return Elements{}, false
	}

	es, _ := e.element.Elements(selector)
	has = len(es) > 0
	elements = make(Elements, 0, len(es))
	for i := range es {
		elements = append(elements, &Element{
			element: es[i],
			page:    e.page,
		})
	}

	return
}

func (e *Element) MustElements(selector string) Elements {
	elements, has := e.Elements(selector)
	if !has {
		panic(&rod.ElementNotFoundError{})
	}

	return elements
}
