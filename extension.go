package browser

import (
	"errors"
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
	for i := range o.Extensions {
		extensionPath := o.Extensions[i]
		path, ok, _, err := o.isExtensionURL(extensionPath)
		if err != nil {
			o.browser.log.Error(err)
			continue
		}

		if ok {
			extensionPath, err = o.downloadExtension(path)
			// if isID && err != nil {
			// 	extensionPath, err = o.downloadExtension("https://statics.ilovechrome.com/crx/download/?id=" + o.Extensions[i])
			// }
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

func (o *Options) downloadExtension(downloadUrl string) (file string, err error) {
	file = zfile.TmpPath() + "/zls-extension/" + zstring.Md5(downloadUrl) + ".crx"
	if zfile.FileSizeUint(file) > 0 {
		return file, nil
	}

	resp, err := zhttp.Get(downloadUrl)
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != 200 {
		return "", errors.New("status code not 200")
	}

	err = resp.ToFile(file)
	if err != nil {
		return "", err
	}

	return file, nil
}

func (o *Options) isExtensionURL(s string) (string, bool, bool, error) {
	if !strings.Contains(s, "/") && !strings.Contains(s, ".") {
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
			return "", false, false, err
		}
		return "https://clients2.google.com/service/update2/crx?response=redirect&prodversion=" + product + "&acceptformat=crx2%2Ccrx3&x=id%3D" + s + "%26uc", true, true, nil
	}

	return s, zvalid.Text(s).IsURL().Ok(), false, nil
}
