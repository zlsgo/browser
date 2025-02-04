package browser

import (
	"time"

	"github.com/go-rod/rod"
)

type Element struct {
	element *rod.Element
	page    *Page
}

type Elements []*Element

func ToElement(e *rod.Element, p *Page) *Element {
	return &Element{
		element: e,
		page:    p,
	}
}

func (e *Element) ROD() *rod.Element {
	return e.element
}

func (e *Element) Timeout(d ...time.Duration) *Element {
	if len(d) > 0 {
		e.element = e.element.Timeout(d[0])
	} else if e.page.timeout != 0 {
		e.element = e.element.Timeout(e.page.timeout)
	}

	return &Element{
		element: e.element,
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
func (e *Element) Element(selector string, jsRegex ...string) (element *Element, has bool) {
	var (
		relm *rod.Element
		err  error
	)
	if len(jsRegex) == 0 {
		relm, err = e.element.Element(selector)
	} else {
		relm, err = e.element.ElementR(selector, jsRegex[0])
	}

	if err != nil {
		return
	}

	return &Element{
		element: relm,
		page:    e.page,
	}, true
}

func (e *Element) MustElement(selector string, jsRegex ...string) *Element {
	elm, has := e.Element(selector, jsRegex...)
	if !has {
		panic(&rod.ElementNotFoundError{})
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
