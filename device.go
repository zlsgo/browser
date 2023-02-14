package browser

import "github.com/go-rod/rod/lib/devices"

type device struct {
}

var Device = device{}

func (d device) IPhoneX() devices.Device {
	return devices.IPhoneX
}

func (d device) IPad() devices.Device {
	return devices.IPad
}

func (d device) IPadMini() devices.Device {
	return devices.IPadMini
}

func (d device) IPadPro() devices.Device {
	return devices.IPadPro
}

func (d device) Pixel2() devices.Device {
	return devices.Pixel2
}

func (d device) Wechat() devices.Device {
	return devices.Device{
		Title:          "Wechat",
		Capabilities:   []string{"touch", "mobile"},
		UserAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 6_1_3 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Mobile/10B329 MicroMessenger/5.0.1",
		AcceptLanguage: "zh-CN,zh;q=0.9",
		Screen: devices.Screen{
			DevicePixelRatio: 2,
			Horizontal: devices.ScreenSize{
				Width:  652,
				Height: 338,
			},
			Vertical: devices.ScreenSize{
				Width:  338,
				Height: 652,
			},
		},
	}
}
