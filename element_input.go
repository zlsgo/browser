package browser

import (
	"github.com/go-rod/rod/lib/input"
)

// FindTextInputElement 查找输入框
func (e *Element) FindTextInputElement(selector ...string) (element *Element, has bool) {
	var s string
	if len(selector) > 0 && selector[0] != "" {
		s = selector[0]
	} else {
		s = "input"
	}

	var elements Elements
	elements, has = e.Elements(s)
	if !has {
		return
	}

	for i := range elements {
		child := elements[i].ROD()
		visible, _ := child.Visible()
		if !visible {
			continue
		}

		typ, err := child.Property("type")
		if err != nil {
			continue
		}

		if typ.String() != "text" && typ.String() != "search" && typ.String() != "textarea" {
			continue
		}
		return &Element{element: child, page: e.page}, true
	}

	return nil, false
}

// InputText 输入文字
func (e *Element) InputText(text string, clear ...bool) error {
	if len(clear) > 0 && clear[0] {
		_ = e.element.SelectAllText()
	}

	return e.element.Input(text)
}

// InputEnter 输入回车
func (e *Element) InputEnter(presskeys ...input.Key) error {
	return e.page.page.KeyActions().Press(presskeys...).Type(input.Enter).Do()
}
