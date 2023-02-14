package browser

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
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
	Hijack *rod.Hijack
}

type HijackProcess func(router *Hijack) bool

func (b *Hijack) BlockDispensable() bool {
	if b.BlockFont() || b.BlockImage() || b.BlockMedia() || b.BlockCSS() || b.BlockFont() || b.BlockPrefetch() {
		return true
	}
	path := b.Hijack.Request.URL().Path
	if path == "/favicon.ico" {
		b.Hijack.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
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
