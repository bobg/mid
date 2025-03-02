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

// JSON produces an http.Handler by JSON encoding and decoding of a given function's input and output.
//
// The signature of the function is:
//
//	func(context.Context, inType) (outType, error)
//
// where inType is any type that can be decoded from JSON
// and outType is any type that can be encoded to JSON.
// These may alternatively be pointers to such types.
//
// Every part of the signature is optional
// (both arguments and both return values).
//
// Passing the wrong type of object to this function produces a panic.
//
// When the function is called:
//
//   - If a context argument is present,
//     it is supplied from the Context() method of the pending *http.Request.
//     That context is further adorned with the pending *http.Request
//     and the pending http.ResponseWriter,
//     which can be retrieved with the Request and ResponseWriter functions.
//
//   - If an inType argument is present,
//     the request is checked to ensure that the method is POST
//     and the Content-Type is application/json;
//     then the request body is unmarshaled into the inType argument.
//     Note that the JSON decoder uses the UseNumber setting;
//     see https://golang.org/pkg/encoding/json/#Decoder.UseNumber.
//
//   - If an outType result is present,
//     it is JSON marshaled and written to the pending ResponseWriter
//     with an HTTP status of 200 (ok).
//     If no outType is present,
//     the default HTTP status is 204 (no content).
//
//   - If an error result is present,
//     it is handled as in Err.
//
// Some of the code in this function is (liberally) adapted from github.com/chain/chain.
func JSON(f interface{}) http.Handler {
	fv := reflect.ValueOf(f)
	if fv.Kind() != reflect.Func {
		jsonPanic(fv.Type())
	}

	ft := fv.Type()
	if ft.IsVariadic() {
		jsonPanic(ft)
	}

	hasCtx, argIsPtr, argType := jsonArgInfo(ft)
	hasErr, hasRes := jsonResultInfo(ft)

	return Err(func(w http.ResponseWriter, req *http.Request) error {
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
			if !strings.EqualFold(req.Method, "POST") {
				return CodeErr{C: http.StatusMethodNotAllowed}
			}

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

			a := argPtr
			if !argIsPtr {
				a = a.Elem()
			}
			args = append(args, a)
		}

		rv := fv.Call(args)

		if hasErr {
			err, _ := rv[len(rv)-1].Interface().(error)
			if err != nil {
				return err
			}
		}

		if !hasRes {
			return nil
		}

		res := rv[0].Interface()
		err := RespondJSON(w, res)
		return errors.Wrap(err, "marshaling JSON response")
	})
}

// RespondJSON responds to an http request with a JSON-encoded object.
func RespondJSON(w http.ResponseWriter, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(obj)
}

func jsonArgInfo(ft reflect.Type) (hasCtx, argIsPtr bool, argType reflect.Type) {
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
		if !ft.In(0).Implements(contextType) {
			jsonPanic(ft)
		}
		hasCtx = true
		argType = ft.In(1)
	default:
		jsonPanic(ft)
	}

	if argType != nil && argType.Kind() == reflect.Ptr {
		argIsPtr = true
		argType = argType.Elem()
	}

	return hasCtx, argIsPtr, argType
}

func jsonResultInfo(ft reflect.Type) (hasErr, hasRes bool) {
	switch ft.NumOut() {
	case 0:
		// do nothing
	case 1:
		if ft.Out(0).Implements(errorType) {
			hasErr = true
		} else {
			hasRes = true
		}
	case 2:
		if !ft.Out(1).Implements(errorType) {
			jsonPanic(ft)
		}
		hasRes = true
		hasErr = true
	default:
		jsonPanic(ft)
	}

	return hasErr, hasRes
}

func jsonPanic(typ reflect.Type) {
	panic(fmt.Sprintf("got %s, want func([ctx], [...]) ([...], [error])", typ))
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
