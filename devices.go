package browser

import (
	"strings"

	"github.com/go-rod/rod/lib/devices"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

type device struct {
}

var Device = device{}

func (d device) NoDefaultDevice() devices.Device {
	return devices.Device{}
}

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

type DeviceOptions struct {
	Name            string
	maxMajorVersion int
	minMajorVersion int
	maxMinorVersion int
	minMinorVersion int
	maxPatchVersion int
	minPatchVersion int
}

func RandomDevice(opt DeviceOptions, device ...devices.Device) devices.Device {
	var d devices.Device
	if len(device) > 0 {
		d = device[0]
	} else {
		d = devices.LaptopWithMDPIScreen
	}

	nameSplit := strings.Split(d.UserAgent, opt.Name+"/")
	if len(nameSplit) < 2 {
		return d
	}

	versionSplit := strings.Split(strings.Split(nameSplit[1], " ")[0], ".")
	originalVersion := strings.Join(versionSplit, ".")

	if opt.maxMajorVersion > 0 {
		if opt.maxMajorVersion == opt.minMajorVersion {
			versionSplit[0] = ztype.ToString(opt.maxMajorVersion)
		} else {
			versionSplit[0] = ztype.ToString(zstring.RandInt(opt.maxMajorVersion, opt.minMajorVersion))
		}
	}

	if opt.maxMinorVersion > 0 {
		if opt.maxMinorVersion == opt.maxMinorVersion {
			versionSplit[1] = ztype.ToString(opt.minMinorVersion)
		} else {
			versionSplit[1] = ztype.ToString(zstring.RandInt(opt.maxMinorVersion, opt.minMinorVersion))
		}
	}

	if opt.maxPatchVersion > 0 {
		if opt.maxPatchVersion == opt.minPatchVersion {
			versionSplit[2] = ztype.ToString(opt.minPatchVersion)
		} else {
			versionSplit[2] = ztype.ToString(zstring.RandInt(opt.maxPatchVersion, opt.minPatchVersion))
		}
	}

	searchValue := opt.Name + "/" + originalVersion
	replaceValue := opt.Name + "/" + strings.Join(versionSplit, ".")

	d.UserAgent = strings.ReplaceAll(d.UserAgent, searchValue, replaceValue)

	return d
}
