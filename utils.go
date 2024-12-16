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
)

var cacheDir = "browser"

func init() {
	launcher.DefaultBrowserDir = filepath.Join(map[string]string{
		"windows": os.Getenv("APPDATA"),
		"darwin":  filepath.Join(os.Getenv("HOME"), ".cache"),
		"linux":   filepath.Join(os.Getenv("HOME"), ".cache"),
	}[runtime.GOOS], cacheDir, "browser")
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

// RandomSleep 随机暂停指定时间范围，单位毫秒
func RandomSleep(ms, maxMS int) {
	time.Sleep(time.Millisecond * time.Duration(ms+rand.Intn(maxMS)))
}
