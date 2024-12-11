package browser

import (
	"errors"
	"time"

	"github.com/go-rod/rod"
)

type Element struct {
	element *rod.Element
	page    *Page
	timeout time.Duration
}

type Elements []*Element

func (e *Element) ROD() *rod.Element {
	return e.element
}

func (e *Element) Timeout(d ...time.Duration) *Element {
	p := &Element{
		element: e.element,
		page:    e.page,
	}
	if len(d) > 0 {
		p.timeout = d[0]
	} else if e.page.Options.Timeout != 0 {
		p.timeout = e.page.Options.Timeout
	} else if e.page.browser.options.Timeout != 0 {
		p.timeout = e.page.browser.options.Timeout
	}

	p.element = e.element.Timeout(p.timeout)
	return p
}

func (e *Element) Element(selector string, jsRegex ...string) (elm *Element, has bool, err error) {
	var relm *rod.Element
	if len(jsRegex) == 0 {
		has, relm, err = e.element.Has(selector)
	} else {
		has, relm, err = e.element.HasR(selector, jsRegex[0])
	}
	if err != nil {
		return
	}
	if !has {
		relm = &rod.Element{}
	}
	elm = &Element{
		element: relm,
		page:    e.page,
	}

	return
}

func (e *Element) MustElement(selector string, jsRegex ...string) *Element {
	elm, has, err := e.Element(selector, jsRegex...)
	if !has {
		err = &rod.ElementNotFoundError{}
	}
	if err != nil {
		panic(err)
	}
	return elm
}

func (e *Element) Elements(selector string) (elems Elements, has bool, err error) {
	var es rod.Elements
	es, err = e.element.Elements(selector)
	if err != nil {
		if errors.Is(err, &rod.ElementNotFoundError{}) {
			return Elements{}, false, err
		}
		return
	}

	has = len(es) > 0
	for _, re := range es {
		elems = append(elems, &Element{
			element: re,
			page:    e.page,
		})
	}

	return
}

func (e *Element) MustElements(selector string) (elems Elements) {
	element, has, err := e.Elements(selector)
	if err != nil {
		panic(err)
	}

	if !has {
		panic(&rod.ElementNotFoundError{})
	}

	return element
}
