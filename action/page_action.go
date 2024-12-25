package action

import (
	"errors"
	"time"

	"github.com/zlsgo/browser"
)

type waitDOMStableType struct {
	timeout time.Duration
	diff    float64
}

var _ ActionType = waitDOMStableType{}

func WaitDOMStable(diff float64, d ...time.Duration) waitDOMStableType {
	o := waitDOMStableType{
		diff: diff,
	}
	if len(d) > 0 {
		o.timeout = d[0]
	}
	return o
}

func (o waitDOMStableType) Do(p *browser.Page) (s any, err error) {
	if o.timeout > 0 {
		return nil, p.WaitDOMStable(o.diff, o.timeout)
	}
	return nil, p.WaitDOMStable(o.diff)
}

func (o waitDOMStableType) Next(p *browser.Page, as Actions, value ActionResult) ([]ActionResult, error) {
	return nil, errors.New("not support")
}
