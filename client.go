package browser

import (
	"github.com/sohaha/zlsgo/zhttp"
)

func (b *Browser) Client() *zhttp.Engine {
	return b.client
}

func (page *Page) Client() *zhttp.Engine {
	return page.browser.client
}
