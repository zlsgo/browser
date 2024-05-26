package browser

import (
	"github.com/ysmood/gson"
)

func (e *Element) Property(name string) (gson.JSON, error) {
	return e.element.Property(name)
}

func (e *Element) MustProperty(name string) gson.JSON {
	return e.element.MustProperty(name)
}

func (e *Element) Text() (string, error) {
	return e.element.Text()
}

func (e *Element) MustText() string {
	return e.element.MustText()
}

func (e *Element) HTML() (string, error) {
	return e.element.HTML()
}

func (e *Element) MustHTML() string {
	return e.element.MustHTML()
}

func (e *Element) HasClassName(className string) bool {
	return e.element.MustEval(`()=>this.classList.contains("` + className + `")`).Bool()
}
