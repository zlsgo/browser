package browser

import (
	"net/http"
	"strings"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zhttp"
)

// Request 发起请求
func (b *Browser) Request(method, url string, v ...interface{}) (*zhttp.Res, error) {
	for _, cookie := range b.cookies {
		v = append(v, cookie)
	}

	resp, err := b.client.Do(method, url, v...)
	if err == nil {
		cookies := zarray.Values(resp.GetCookie())
		b.SetCookies(cookies)
	}

	return resp, err
}

// Request 发起请求
func (page *Page) Request(method, url string, v ...interface{}) (*zhttp.Res, error) {
	return page.browser.Request(method, url, v...)
}

// SavePageCookie 保存页面 cookie
func (page *Page) SavePageCookie() (cookies []*http.Cookie) {
	for _, cookie := range page.page.MustCookies() {
		cookies = append(cookies, &http.Cookie{
			Name:   cookie.Name,
			Value:  strings.Trim(cookie.Value, "\""),
			Path:   cookie.Path,
			Domain: cookie.Domain,
		})
	}

	page.browser.cookies = page.browser.uniqueCookies(cookies)

	return page.browser.cookies
}
