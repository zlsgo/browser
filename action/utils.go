package action

import (
	"github.com/zlsgo/browser"
)

// ExtractElement 从 Result 里获取元素
func ExtractElement(parentResults ...ActionResult) (*browser.Element, bool) {
	if len(parentResults) == 0 || parentResults[0].Value == nil {
		return nil, false
	}
	switch v := parentResults[0].Value.(type) {
	case *browser.Element:
		return v, true
	}

	return nil, false
}

// ExtractPage 从 Result 里获取页面
func ExtractPage(parentResults ...ActionResult) (*browser.Page, bool) {
	if len(parentResults) == 0 || parentResults[0].Value == nil {
		return nil, false
	}
	switch v := parentResults[0].Value.(type) {
	case *browser.Page:
		return v, true
	}
	return nil, false
}
