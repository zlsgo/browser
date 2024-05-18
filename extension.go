package browser

import (
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/mediabuyerbot/go-crx3"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zvalid"
)

func (o *Options) handerExtension() (extensions []string) {
	for _, extensionPath := range o.Extensions {
		path, ok, err := o.isExtensionURL(extensionPath)
		if err != nil {
			o.browser.log.Error(err)
			continue
		}

		if ok {
			extensionPath, err = o.downloadExtension(path)
			if err != nil {
				o.browser.log.Error(err)
				continue
			}
		}

		if zfile.FileExist(extensionPath) && strings.EqualFold(filepath.Ext(extensionPath), ".crx") {
			dir := extensionPath[:len(extensionPath)-4]
			if !zfile.DirExist(dir) {
				_ = crx3.Extension(extensionPath).Unpack()
			}
			extensionPath = dir
		}

		if zfile.DirExist(extensionPath) {
			extensions = append(extensions, extensionPath)
		}
	}

	return
}

func (o *Options) downloadExtension(downloadUrl string) (string, error) {
	file := zfile.TmpPath() + "/zls-extension/" + zstring.Md5(downloadUrl) + ".crx"
	if zfile.FileExist(file) {
		return file, nil
	}

	resp, err := zhttp.Get(downloadUrl)
	if err != nil {
		return "", err
	}

	err = resp.ToFile(file)
	if err != nil {
		return "", err
	}

	return file, nil
}

func (o *Options) isExtensionURL(s string) (string, bool, error) {
	if !strings.Contains(s, "/") {
		var product string
		err := zerror.TryCatch(func() error {
			browser := rod.New().ControlURL(launcher.New().Bin(getBin(o.Bin)).MustLaunch()).MustConnect()
			vResult, err := browser.Version()
			if err == nil {
				product = ztype.ToString(vResult.Product[1])
			}
			go browser.Close()
			return err
		})
		if err != nil {
			return "", false, err
		}

		return "https://clients2.google.com/service/update2/crx?response=redirect&prodversion=" + product + "&acceptformat=crx2%2Ccrx3&x=id%3D" + s + "%26uc", true, nil
	}

	return s, zvalid.Text(s).IsURL().Ok(), nil
}
