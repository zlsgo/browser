package browser

import (
	"time"

	"github.com/go-rod/rod/lib/devices"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zlog"
)

// BrowserBuilder 浏览器构建器，提供流式配置API
type BrowserBuilder struct {
	options Options
}

// Preset 预设配置类型
type Preset string

const (
	PresetDevelopment Preset = "development"
	PresetProduction  Preset = "production"
	PresetTesting     Preset = "testing"
	PresetStealth     Preset = "stealth"
)

// NewBrowser 创建新的浏览器构建器
func NewBrowser() *BrowserBuilder {
	return &BrowserBuilder{
		options: Options{
			autoKill: true,
			Headless: true,
			Flags: map[string]string{
				"no-sandbox":               "",
				"disable-blink-features":   "AutomationControlled",
				"no-default-browser-check": "",
				"no-first-run":             "",
				"disable-component-update": "",
				"window-position":          "0,0",
			},
			IgnoreCertError: true,
		},
	}
}

// Preset 使用预设配置
func (b *BrowserBuilder) Preset(preset Preset) *BrowserBuilder {
	switch preset {
	case PresetDevelopment:
		b.options.Headless = false
		b.options.Devtools = true
		b.options.Debug = true
		b.options.SlowMotion = 100 * time.Millisecond
	case PresetProduction:
		b.options.Headless = true
		b.options.Leakless = true
		b.options.Stealth = true
	case PresetTesting:
		b.options.Headless = true
		b.options.Timeout = 30 * time.Second
		b.options.Incognito = true
	case PresetStealth:
		b.options.Headless = true
		b.options.Stealth = true
		b.options.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}
	return b
}

// WithHeadless 设置无头模式
func (b *BrowserBuilder) WithHeadless(headless bool) *BrowserBuilder {
	b.options.Headless = headless
	return b
}

// WithUserAgent 设置用户代理
func (b *BrowserBuilder) WithUserAgent(userAgent string) *BrowserBuilder {
	b.options.UserAgent = userAgent
	return b
}

// WithTimeout 设置超时时间
func (b *BrowserBuilder) WithTimeout(timeout time.Duration) *BrowserBuilder {
	b.options.Timeout = timeout
	return b
}

// WithProxy 设置代理
func (b *BrowserBuilder) WithProxy(proxyURL string) *BrowserBuilder {
	b.options.ProxyUrl = proxyURL
	return b
}

// WithDevice 设置设备模拟
func (b *BrowserBuilder) WithDevice(device devices.Device) *BrowserBuilder {
	b.options.DefaultDevice = device
	return b
}

// WithUserDataDir 设置用户数据目录
func (b *BrowserBuilder) WithUserDataDir(dir string) *BrowserBuilder {
	b.options.UserDataDir = dir
	return b
}

// WithFlag 添加启动标志
func (b *BrowserBuilder) WithFlag(flag, value string) *BrowserBuilder {
	if b.options.Flags == nil {
		b.options.Flags = make(map[string]string)
	}
	b.options.Flags[flag] = value
	return b
}

// WithExtension 添加扩展
func (b *BrowserBuilder) WithExtension(extensionPath string) *BrowserBuilder {
	b.options.Extensions = append(b.options.Extensions, extensionPath)
	return b
}

// WithScript 添加启动脚本
func (b *BrowserBuilder) WithScript(script string) *BrowserBuilder {
	b.options.Scripts = append(b.options.Scripts, script)
	return b
}

// WithIncognito 设置隐身模式
func (b *BrowserBuilder) WithIncognito(incognito bool) *BrowserBuilder {
	b.options.Incognito = incognito
	return b
}

// WithStealth 设置隐形模式（反检测）
func (b *BrowserBuilder) WithStealth(stealth bool) *BrowserBuilder {
	b.options.Stealth = stealth
	return b
}

// WithDevtools 开启开发者工具
func (b *BrowserBuilder) WithDevtools(devtools bool) *BrowserBuilder {
	b.options.Devtools = devtools
	return b
}

// WithUserMode 使用用户模式（使用现有浏览器）
func (b *BrowserBuilder) WithUserMode(userMode bool) *BrowserBuilder {
	b.options.UserMode = userMode
	return b
}

// WithSlowMotion 设置慢动作延迟
func (b *BrowserBuilder) WithSlowMotion(delay time.Duration) *BrowserBuilder {
	b.options.SlowMotion = delay
	return b
}

// WithLanguage 设置接受语言
func (b *BrowserBuilder) WithLanguage(lang string) *BrowserBuilder {
	b.options.AcceptLanguage = lang
	return b
}

// WithBin 设置浏览器二进制路径
func (b *BrowserBuilder) WithBin(binPath string) *BrowserBuilder {
	b.options.Bin = binPath
	return b
}

// WithWSEndpoint 设置WebSocket端点
func (b *BrowserBuilder) WithWSEndpoint(endpoint string) *BrowserBuilder {
	b.options.WSEndpoint = endpoint
	return b
}

// Build 构建浏览器实例
func (b *BrowserBuilder) Build() (*Browser, error) {
	browser := &Browser{
		client:  zhttp.New(),
		log:     zlog.New(),
		options: b.options,
	}

	browser.options.browser = browser
	browser.Client().EnableCookie(true)
	browser.canUserDir = browser.options.UserMode || browser.options.UserDataDir != ""

	if err := browser.init(); err != nil {
		return nil, err
	}

	return browser, nil
}

// MustBuild 构建浏览器实例，失败时panic
func (b *BrowserBuilder) MustBuild() *Browser {
	browser, err := b.Build()
	if err != nil {
		panic(err)
	}
	return browser
}

// BuildAndConnect 构建并连接浏览器
func (b *BrowserBuilder) BuildAndConnect() (*Browser, error) {
	browser, err := b.Build()
	if err != nil {
		return nil, err
	}
	return browser, nil
}

// Clone 克隆构建器，用于创建相似配置的多个实例
func (b *BrowserBuilder) Clone() *BrowserBuilder {
	newBuilder := *b

	// 深拷贝maps和slices
	if b.options.Flags != nil {
		newBuilder.options.Flags = make(map[string]string)
		for k, v := range b.options.Flags {
			newBuilder.options.Flags[k] = v
		}
	}

	if b.options.Extensions != nil {
		newBuilder.options.Extensions = make([]string, len(b.options.Extensions))
		copy(newBuilder.options.Extensions, b.options.Extensions)
	}

	if b.options.Scripts != nil {
		newBuilder.options.Scripts = make([]string, len(b.options.Scripts))
		copy(newBuilder.options.Scripts, b.options.Scripts)
	}

	if b.options.Envs != nil {
		newBuilder.options.Envs = make([]string, len(b.options.Envs))
		copy(newBuilder.options.Envs, b.options.Envs)
	}

	return &newBuilder
}

// 常用设备预设
var (
	// DeviceDesktop 桌面设备
	DeviceDesktop = devices.LaptopWithHiDPIScreen

	// DeviceMobile 移动设备
	DeviceMobile = devices.IPhoneX

	// DeviceTablet 平板设备
	DeviceTablet = devices.IPad
)

// WithMobileDevice 设置移动设备模拟
func (b *BrowserBuilder) WithMobileDevice() *BrowserBuilder {
	return b.WithDevice(DeviceMobile)
}

// WithTabletDevice 设置平板设备模拟
func (b *BrowserBuilder) WithTabletDevice() *BrowserBuilder {
	return b.WithDevice(DeviceTablet)
}

// WithDesktopDevice 设置桌面设备模拟
func (b *BrowserBuilder) WithDesktopDevice() *BrowserBuilder {
	return b.WithDevice(DeviceDesktop)
}
