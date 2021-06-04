// Package mid contains assorted middleware for use in HTTP services.
package mid

import (
	"net/http"
)

// ResponseWrapper implements http.ResponseWriter,
// delegating calls to a wrapped http.ResponseWriter object.
// It also records the status code and the number of response bytes that have been written.
type ResponseWrapper struct {
	// W is the wrapped ResponseWriter to which method calls are delegated.
	W http.ResponseWriter

	// N is the number of bytes that have been written with calls to Write.
	N int

	// Code is the status code that has been written with WriteHeader,
	// or zero if no call to WriteHeader has yet been made.
	// If Write is called before any call to WriteHeader,
	// then this is set to http.StatusOK (200).
	Code int
}

// Header implements http.ResponseWriter.Header.
func (ww *ResponseWrapper) Header() http.Header {
	return ww.W.Header()
}

// Write implements http.ResponseWriter.Write.
func (ww *ResponseWrapper) Write(b []byte) (int, error) {
	if ww.Code == 0 {
		ww.Code = http.StatusOK
	}
	n, err := ww.W.Write(b)
	ww.N += n
	return n, err
}

// WriteHeader implements http.ResponseWriter.WriteHeader.
func (ww *ResponseWrapper) WriteHeader(code int) {
	ww.Code = code
	ww.W.WriteHeader(code)
}
