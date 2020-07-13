// Package mid contains assorted middleware for use in HTTP services.
package mid

import (
	"net/http"
)

type responseWriterWrapper struct {
	w    http.ResponseWriter
	n    int
	code int
}

func (ww *responseWriterWrapper) Header() http.Header {
	return ww.w.Header()
}

func (ww *responseWriterWrapper) Write(b []byte) (int, error) {
	if ww.code == 0 {
		ww.code = http.StatusOK
	}
	n, err := ww.w.Write(b)
	ww.n += n
	return n, err
}

func (ww *responseWriterWrapper) WriteHeader(code int) {
	ww.code = code
	ww.w.WriteHeader(code)
}
