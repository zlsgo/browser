package browser

import (
	"net/http"
	"strings"

	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zhttp"
)

func (b *Browser) SetCookie(page *Page) (cookies []*http.Cookie) {
	for _, cookie := range page.page.MustCookies() {
		cookies = append(cookies, &http.Cookie{
			Name:   cookie.Name,
			Value:  strings.Trim(cookie.Value, "\""),
			Path:   cookie.Path,
			Domain: cookie.Domain,
		})
	}

	b.cookies = cookies

	return
}

func (b *Browser) setUserAgent(page *Page) *proto.NetworkSetUserAgentOverride {
	if b.userAgent == nil {
		b.userAgent = &proto.NetworkSetUserAgentOverride{}
	}

	resp, err := page.page.Eval(`() => navigator.userAgent`)
	if err == nil {
		b.userAgent.UserAgent = resp.Value.String()
	}

	resp, err = page.page.Eval(`() => navigator.language`)
	if err == nil {
		b.userAgent.AcceptLanguage = resp.Value.String()
	}

	return b.userAgent
}

func (b *Browser) Request(method, url string, v ...interface{}) (*zhttp.Res, error) {
	for _, cookie := range b.cookies {
		v = append(v, cookie)
	}

	resp, err := b.client.Do(method, url, v...)

	return resp, err
}

func (page *Page) Request(method, url string, v ...interface{}) (*zhttp.Res, error) {
	return page.browser.Request(method, url, v...)
}

func (page *Page) SetCookie() (cookies []*http.Cookie) {
	return page.browser.SetCookie(page)
}
