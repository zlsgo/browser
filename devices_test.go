package browser

import (
	"testing"

	"github.com/go-rod/rod/lib/devices"
	"github.com/sohaha/zlsgo"
)

func TestRandomDevice(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Log(devices.Nexus4.UserAgent)
	tt.Log(RandomDevice(devices.Nexus4, DeviceOptions{
		Name:            "Chrome",
		maxMajorVersion: 112,
		minMajorVersion: 100,
		maxMinorVersion: 20,
		maxPatchVersion: 5000,
	}).UserAgent)
}
