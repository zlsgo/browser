package browser

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-rod/rod/lib/proto"
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
		b.SetPageCookie(cookies)
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

	page.browser.cookies = append(page.browser.cookies, cookies...)

	return cookies
}

// SetPageCookie 设置页面 cookie
func (b *Browser) SetPageCookie(cookies []*http.Cookie) error {
	if cookies == nil {
		b.cookies = make([]*http.Cookie, 0, 0)
		_ = b.Browser.SetCookies(nil)
		return nil
	}

	c, err := b.cookiesToProto(cookies)
	if err != nil {
		return err
	}

	b.Browser.SetCookies(c)
	return nil
}

func (b *Browser) cookiesToProto(cookies []*http.Cookie) ([]*proto.NetworkCookieParam, error) {
	protoCookies := make([]*proto.NetworkCookieParam, 0, len(cookies))
	for i := range cookies {
		if cookies[i].Domain == "" {
			return nil, errors.New("domain is required for cookie configuration")
		}
		protoCookies = append(protoCookies, &proto.NetworkCookieParam{
			Name:     cookies[i].Name,
			Value:    cookies[i].Value,
			Expires:  proto.TimeSinceEpoch(cookies[i].Expires.Unix()),
			Path:     cookies[i].Path,
			Domain:   cookies[i].Domain,
			Secure:   cookies[i].Secure,
			HTTPOnly: cookies[i].HttpOnly,
		})
	}

	return protoCookies, nil
}
