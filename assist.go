package browser

import (
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zfile"
)

// ScreenshotFullPage 截图全屏
func (p *Page) ScreenshotFullPage(file string) error {
	b, err := p.ROD().Screenshot(true, nil)
	if err != nil {
		return err
	}

	return zfile.WriteFile(zfile.RealPath(file), b)
}

// Screenshot 截图
func (p *Page) Screenshot(file string) error {
	b, err := p.ROD().Screenshot(false, nil)
	if err != nil {
		return err
	}

	return zfile.WriteFile(zfile.RealPath(file), b)
}

// Screenshot 截图元素
func (ele *Element) Screenshot(file string) error {
	b, err := ele.ROD().Screenshot(proto.PageCaptureScreenshotFormatPng, 0)
	if err != nil {
		return err
	}

	return zfile.WriteFile(zfile.RealPath(file), b)
}
