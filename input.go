package browser

import (
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zstring"
)

func (page *Page) MouseMove(x, y float64, steps ...float64) error {
	to := proto.Point{X: x, Y: y}

	if len(steps) == 0 {
		return page.page.Mouse.MoveTo(to)
	}

	return page.page.Mouse.MoveLinear(to, int(steps[0]))
}

func (page *Page) MouseMoveToElement(ele *Element, steps ...float64) error {
	shape, err := ele.element.Shape()
	if err != nil {
		return err
	}

	box := shape.Box()

	to := proto.Point{X: box.X + float64(zstring.RandInt(0, int(box.Width))), Y: box.Y + float64(zstring.RandInt(0, int(box.Height)))}

	if len(steps) == 0 {
		return page.page.Mouse.MoveTo(to)
	}

	return page.page.Mouse.MoveLinear(to, int(steps[0]))
}
