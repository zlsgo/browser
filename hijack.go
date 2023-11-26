package browser

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zhttp"
)

type HijackRouter struct {
	router *rod.HijackRouter
}

func (h *HijackRouter) Add(pattern string, handler func(*Hijack) bool) {
	h.router.Add(pattern, "", func(h *rod.Hijack) {
		handler(&Hijack{h})
	})
}

type Hijack struct {
	*rod.Hijack
}

func NewHijackRouter(fn func(b *Hijack) bool) map[string]HijackProcess {
	return map[string]HijackProcess{
		"*": fn,
	}
}

type HijackProcess func(router *Hijack) bool

type HijackData struct {
	URL            *url.URL
	Request        []byte
	Response       []byte
	Method         string
	StatusCode     int
	Header         http.Header
	ResponseHeader http.Header
}

func (h *Hijack) HijackRequests(fn func(d *HijackData, err error) bool) bool {
	if h.Hijack.Request.Req() != nil && h.Hijack.Request.Req().URL != nil {
		data := &HijackData{
			URL:    h.Hijack.Request.Req().URL,
			Header: h.Hijack.Request.Req().Header,
			Method: h.Hijack.Request.Method(),
		}
		// reqBytes, _ = httputil.DumpRequest(h.Hijack.Request.Req(), true)

		if h.Hijack.Request.Method() == http.MethodPost {
			var save, body io.ReadCloser
			save, body, _ = copyBody(h.Hijack.Request.Req().Body)
			data.Request, _ = ioutil.ReadAll(save)
			h.Hijack.Request.Req().Body = body
		}

		err := h.Hijack.LoadResponse(zhttp.Client(), true)
		if err == nil {
			data.StatusCode = h.Hijack.Response.Payload().ResponseCode
			data.Response = h.Hijack.Response.Payload().Body
			data.ResponseHeader = transformHeaders(h.Hijack.Response.Payload().ResponseHeaders)
		}

		return fn(data, err)
	}

	return false
}

func (h *Hijack) BlockDispensable() bool {
	if h.BlockFont() || h.BlockImage() || h.BlockMedia() || h.BlockCSS() || h.BlockFont() || h.BlockPrefetch() || h.BlockFavicon() {
		return true
	}

	return false
}

func (h *Hijack) BlockFavicon() bool {
	path := h.Hijack.Request.URL().Path
	if path == "/favicon.ico" {
		h.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return true
	}

	return false
}

func (h *Hijack) BlockFont() bool {
	if h.Hijack.Request.Type() == proto.NetworkResourceTypeFont {
		h.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return true
	}

	return false
}

func (h *Hijack) BlockPrefetch() bool {
	if h.Hijack.Request.Type() == proto.NetworkResourceTypePrefetch {
		h.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return true
	}

	return false
}

func (h *Hijack) BlockMedia() bool {
	if h.Hijack.Request.Type() == proto.NetworkResourceTypeMedia {
		h.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return true
	}
	return false
}

func (h *Hijack) BlockJS() bool {
	if h.Hijack.Request.Type() == proto.NetworkResourceTypeScript {
		h.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return true
	}
	return false
}

func (h *Hijack) BlockCSS() bool {
	if h.Hijack.Request.Type() == proto.NetworkResourceTypeStylesheet {
		h.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return true
	}
	return false
}

func (h *Hijack) BlockImage() bool {
	if h.Hijack.Request.Type() == proto.NetworkResourceTypeImage {
		h.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		return true
	}
	return false
}
