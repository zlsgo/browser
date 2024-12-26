package action

import (
	"time"

	"github.com/sohaha/zlsgo/zjson"
	"github.com/zlsgo/browser"
)

func NewAutoFromJson(b *browser.Browser, jsonStr []byte) (*Auto, error) {
	j := zjson.ParseBytes(jsonStr)
	url := j.Get("url").String()
	actions := j.Get("actions").Array()
	timeout := j.Get("timeout").Int()

	return &Auto{
		browser: b,
		url:     url,
		actions: parseAction(actions),
		timeout: time.Duration(timeout) * time.Second,
	}, nil
}
