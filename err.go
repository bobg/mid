package mid

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

// Errf is a convenience wrapper for http.Error.
// It calls http.Error(w, fmt.Sprintf(format, args...), code).
// It also logs that message with log.Print.
// If code is 0, it defaults to http.StatusInternalServerError.
// If format is "", Errf uses http.StatusText instead.
func Errf(w http.ResponseWriter, code int, format string, args ...interface{}) {
	if code == 0 {
		code = http.StatusInternalServerError
	}

	var msg string
	if format == "" {
		msg = http.StatusText(code)
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	log.Print(msg)
	http.Error(w, msg, code)
}

// Err wraps an error-returning function as an http.Handler.
// If the returned error is a Responder (such as a CodeErr),
// its Respond method is used to respond to the request.
// Otherwise, an error return will set the response code to http.StatusInternalServerError
// (if a response code has not already been set) with the message set to the error's string.
// If there is no error, the status code (if not already set)
// defaults to http.StatusOK, or http.StatusNoContent if nothing has been written to w.
func Err(f func(http.ResponseWriter, *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ww := responseWriterWrapper{w: w}
		err := f(&ww, req)
		var responder Responder
		if errors.As(err, &responder) {
			responder.Respond(w)
		} else if err != nil {
			if ww.code == 0 {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else if ww.code == 0 {
			code := http.StatusOK
			if ww.n == 0 {
				code = http.StatusNoContent
			}
			w.WriteHeader(code)
		}
	})
}

// CodeErr is an error that can be returned from the function wrapped by Err
// to control the HTTP status code returned from the pending request.
type CodeErr struct {
	// C is an HTTP status code.
	C int

	// Err is an optional wrapped error.
	Err error
}

// Error implements the error interface.
func (c CodeErr) Error() string {
	s := fmt.Sprintf("HTTP %d", c.C)
	if t := http.StatusText(c.C); t != "" {
		s += ": " + t
	}
	if c.Err != nil {
		s += ": " + c.Err.Error()
	}
	return s
}

// Unwrap implements the interface for errors.Unwrap.
func (c CodeErr) Unwrap() error {
	return c.Err
}

// As implements the interface for errors.As.
func (c CodeErr) As(target interface{}) bool {
	if ptr, ok := target.(*CodeErr); ok {
		*ptr = c
		return true
	}
	return false
}

// Respond implements Responder.
func (c CodeErr) Respond(w http.ResponseWriter) {
	http.Error(w, c.Error(), c.C)
}

// Responder is an interface for objects that know how to respond to an HTTP request.
// It is useful in the case of errors that want to set custom error strings and/or status codes
// (e.g. via http.Error).
type Responder interface {
	Respond(http.ResponseWriter)
}
