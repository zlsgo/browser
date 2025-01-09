package action

import (
	"errors"
	"time"

	"github.com/sohaha/zlsgo/zjson"
	"github.com/zlsgo/browser"
)

func NewAutoFromJson(b *browser.Browser, jsonStr []byte) (*Auto, error) {
	j := zjson.ParseBytes(jsonStr)
	url := j.Get("url").String()
	actions := j.Get("actions").Array()
	timeout := j.Get("timeout").Int()

	if url == "" {
		return nil, errors.New("url is required")
	}

	if len(actions) == 0 {
		return nil, errors.New("actions is required")
	}

	return &Auto{
		browser: b,
		url:     url,
		actions: parseAction(actions),
		timeout: time.Duration(timeout) * time.Second,
	}, nil
}
