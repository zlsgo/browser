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

type Hijack struct {
	*rod.Hijack
	client *zhttp.Engine
	abort  bool
}

func (h *Hijack) Abort() {
	h.abort = true
}

func newHijacl(h *rod.Hijack, client *zhttp.Engine) *Hijack {
	return &Hijack{
		client: client,
		Hijack: h,
	}
}

func HijackAllRouter(fn func(b *Hijack) (stop bool)) map[string]HijackProcess {
	return map[string]HijackProcess{
		"*": fn,
	}
}

type HijackProcess func(router *Hijack) (stop bool)

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

		err := h.Hijack.LoadResponse(h.client.Client(), true)
		if err == nil {
			data.StatusCode = h.Hijack.Response.Payload().ResponseCode
			data.Response = h.Hijack.Response.Payload().Body
			data.ResponseHeader = transformHeaders(h.Hijack.Response.Payload().ResponseHeaders)
		}

		return fn(data, err)
	}

	return false
}

func (h *Hijack) IsDispensable() bool {
	return h.IsFont() || h.IsImage() || h.IsMedia() || h.IsCSS() || h.IsFont() || h.IsPrefetch() || h.IsFavicon()
}

func (h *Hijack) IsFavicon() bool {
	return h.Hijack.Request.URL().Path == "/favicon.ico"
}

func (h *Hijack) IsFont() bool {
	return h.Hijack.Request.Type() == proto.NetworkResourceTypeFont
}

func (h *Hijack) IsPrefetch() bool {
	return h.Hijack.Request.Type() == proto.NetworkResourceTypePrefetch
}

func (h *Hijack) IsMedia() bool {
	return h.Hijack.Request.Type() == proto.NetworkResourceTypeMedia
}

func (h *Hijack) IsJS() bool {
	return h.Hijack.Request.Type() == proto.NetworkResourceTypeScript
}

func (h *Hijack) IsCSS() bool {
	return h.Hijack.Request.Type() == proto.NetworkResourceTypeStylesheet
}

func (h *Hijack) IsImage() bool {
	return h.Hijack.Request.Type() == proto.NetworkResourceTypeImage
}
