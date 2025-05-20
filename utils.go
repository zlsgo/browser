package browser

import (
	"bytes"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zutil"
)

var cacheDir = "browser"

func init() {
	launcher.DefaultBrowserDir = filepath.Join(map[string]string{
		"windows": os.Getenv("APPDATA"),
		"darwin":  filepath.Join(os.Getenv("HOME"), ".cache"),
		"linux":   filepath.Join(os.Getenv("HOME"), ".cache"),
	}[runtime.GOOS], cacheDir, "browser")
}

func isDebian() bool {
	if !zutil.IsLinux() {
		return false
	}

	resp, _ := zfile.ReadFile("/etc/os-release")
	if len(resp) == 0 {
		return false
	}
	return bytes.Contains(resp, []byte("debian"))
}

func copyBody(b io.ReadCloser) (io.ReadCloser, io.ReadCloser, error) {
	if b == nil || b == http.NoBody {
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err := b.Close(); err != nil {
		return nil, b, err
	}
	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func transformHeaders(h []*proto.FetchHeaderEntry) http.Header {
	newHeader := http.Header{}
	for _, data := range h {
		newHeader.Add(data.Name, data.Value)
	}
	return newHeader
}

// RandomSleep randomly pause for a specified time range, unit in milliseconds
func RandomSleep(ms, maxMS int) {
	time.Sleep(time.Millisecond * time.Duration(ms+rand.Intn(maxMS)))
}

// uniqueCookies duplicate removal processing for cookies
// When encountering a Cookie with the same combination of Name+Path+Domain, the latter Cookie will overwrite the former Cookie
func (browser *Browser) uniqueCookies(cookies []*http.Cookie) []*http.Cookie {
	cookieMap := make(map[string]*http.Cookie)

	for _, c := range browser.cookies {
		key := c.Name + c.Path + c.Domain
		cookieMap[key] = c
	}

	for _, c := range cookies {
		key := c.Name + c.Path + c.Domain
		cookieMap[key] = c
	}

	nCookies := make([]*http.Cookie, 0, len(cookieMap))
	for i := range cookieMap {
		nCookies = append(nCookies, cookieMap[i])
	}

	return nCookies
}
