package browser

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-rod/rod/lib/proto"
)

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
