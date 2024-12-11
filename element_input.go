package browser

import (
	"errors"

	"github.com/go-rod/rod/lib/input"
)

// FindTextInputElement  find text input element
func (e *Element) FindTextInputElement(selector ...string) (elm *Element, has bool, err error) {
	var s string
	if len(selector) > 0 {
		s = selector[0]
	} else {
		s = "input"
	}

	var elms Elements
	elms, has, err = e.Elements(s)
	if err != nil || !has {
		return
	}

	for i := range elms {
		child := elms[i].ROD()
		visible, _ := child.Visible()
		if !visible {
			continue
		}

		typ, err := child.Property("type")
		if err != nil {
			continue
		}

		if typ.String() != "text" && typ.String() != "search" {
			continue
		}
		return &Element{element: child, page: e.page}, true, nil
	}

	return nil, false, errors.New("not found")
}

func (e *Element) InputText(text string, clear ...bool) error {
	if len(clear) > 0 && clear[0] {
		_ = e.element.SelectAllText()
	}

	return e.element.Input(text)
}

func (e *Element) InputEnter(presskeys ...input.Key) error {
	return e.page.page.KeyActions().Press(presskeys...).Type(input.Enter).Do()
}
