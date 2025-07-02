package browser

import (
	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"
)

// Box 获取元素的边界框
func (e *Element) Box() (*proto.DOMRect, error) {
	shape, err := e.element.Shape()
	if err != nil {
		return nil, err
	}
	return shape.Box(), nil
}

// MustBox 获取元素的边界框，如果出错则 panic
func (e *Element) MustBox() *proto.DOMRect {
	box, err := e.Box()
	if err != nil {
		panic(err)
	}
	return box
}

// Property 获取元素的属性值
func (e *Element) Property(name string) (gson.JSON, error) {
	return e.element.Property(name)
}

// MustProperty 获取元素的属性值，如果出错则 panic
func (e *Element) MustProperty(name string) gson.JSON {
	return e.element.MustProperty(name)
}

// Text 获取元素的文本内容
func (e *Element) Text() (string, error) {
	return e.element.Text()
}

// MustText 获取元素的文本内容，如果出错则 panic
func (e *Element) MustText() string {
	return e.element.MustText()
}

// HTML 获取元素的 HTML 内容
func (e *Element) HTML() (string, error) {
	return e.element.HTML()
}

// MustHTML 获取元素的 HTML 内容，如果出错则 panic
func (e *Element) MustHTML() string {
	return e.element.MustHTML()
}

// HasClassName 检查元素是否包含指定的类名
func (e *Element) HasClassName(className string) bool {
	return e.element.MustEval(`()=>this.classList.contains("` + className + `")`).Bool()
}

// TagName 获取元素的标签名（小写）
func (e *Element) TagName() (string, error) {
	result, err := e.element.Eval(`() => this.tagName.toLowerCase()`)
	if err != nil {
		return "", err
	}
	return result.Value.String(), nil
}

// MustTagName 获取元素的标签名（小写）
func (e *Element) MustTagName() string {
	result, err := e.TagName()
	if err != nil {
		panic(err)
	}
	return result
}
