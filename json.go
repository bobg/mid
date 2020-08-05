package mid

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

type (
	reqKey  struct{}
	respKey struct{}
)

// JSON produces an http.Handler from a function f.
//
// The function may take one "request" argument of any type that can be JSON-unmarshaled.
// That argument can optionally be preceded by a context.Context.
// It may return one "response" value of any type that can be JSON-marshaled.
// That return value can optionally be followed by an error.
// If the function returns an error, the error is handled as in Err.
//
// When the handler is invoked,
// the request is checked to ensure that the method is POST
// and the Content-Type is application/json.
// Then the function f is invoked with the request body JSON-unmarshaled into its argument
// (if there is one).
// The return value of f (if there is one) is JSON-marshaled into the response
// and the Content-Type of the response is set to application/json.
//
// If f takes a context.Context, it receives the context object from the http.Request.
// The context object is adorned with the pending *http.Request,
// which can be retrieved with the Request function,
// and the pending http.ResponseWriter,
// which can be retrieved with ResponseWriter.
//
// Some of the code in this function is (liberally) adapted from github.com/chain/chain.
func JSON(f interface{}) http.Handler {
	e := func() string { return fmt.Sprintf("got %T, want func([ctx], [...]) ([...], [error])", f) }

	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		panic(e())
	}

	ft := fv.Type()
	if ft.IsVariadic() {
		panic(e())
	}

	var (
		hasCtx  bool
		argType reflect.Type
	)

	switch ft.NumIn() {
	case 0:
		// do nothing
	case 1:
		if ft.In(0).Implements(contextType) {
			hasCtx = true
		} else {
			argType = ft.In(0)
		}
	case 2:
		if ft.In(0).Implements(contextType) {
			hasCtx = true
		} else {
			panic(e())
		}
		argType = ft.In(1)
	default:
		panic(e())
	}

	var (
		hasErr     bool
		resultType reflect.Type
	)

	switch ft.NumOut() {
	case 0:
		// do nothing
	case 1:
		if ft.Out(0).Implements(errorType) {
			hasErr = true
		} else {
			resultType = ft.Out(0)
		}
	case 2:
		if ft.Out(1).Implements(errorType) {
			hasErr = true
		} else {
			panic(e())
		}
		resultType = ft.Out(0)
	default:
		panic(e())
	}

	return Err(func(w http.ResponseWriter, req *http.Request) error {
		if !strings.EqualFold(req.Method, "POST") {
			return CodeErr{C: http.StatusMethodNotAllowed}
		}

		ctx := req.Context()
		if hasCtx {
			ctx = context.WithValue(ctx, reqKey{}, req)
			ctx = context.WithValue(ctx, respKey{}, w)
		}

		var args []reflect.Value
		if hasCtx {
			args = append(args, reflect.ValueOf(ctx))
		}
		if argType != nil {
			ctfield := req.Header.Get("Content-Type")
			ct, _, err := mime.ParseMediaType(ctfield)
			if err != nil {
				return CodeErr{C: http.StatusBadRequest, Err: err}
			}
			if !strings.EqualFold(ct, "application/json") {
				return CodeErr{C: http.StatusBadRequest}
			}

			argPtr := reflect.New(argType)
			dec := json.NewDecoder(req.Body)
			dec.UseNumber()
			err = dec.Decode(argPtr.Interface())
			if err != nil {
				return errors.Wrap(err, "unmarshaling JSON argument")
			}
			args = append(args, argPtr.Elem())
		}

		rv := fv.Call(args)

		if hasErr {
			err, _ := rv[len(rv)-1].Interface().(error)
			if err != nil {
				return err
			}
		}

		if resultType == nil {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}

		res := rv[0].Interface()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := json.NewEncoder(w).Encode(res)
		return errors.Wrap(err, "marshaling JSON response")
	})
}

// Request returns the pending *http.Request object
// when called on the context passed to a JSON handler.
func Request(ctx context.Context) *http.Request {
	req, _ := ctx.Value(reqKey{}).(*http.Request)
	return req
}

// ResponseWriter returns the pending http.ResponseWriter object
// when called on the context passed to a JSON handler.
func ResponseWriter(ctx context.Context) http.ResponseWriter {
	resp, _ := ctx.Value(respKey{}).(http.ResponseWriter)
	return resp
}

func panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
