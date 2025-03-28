package mid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/bobg/errors"
)

type traceIDKeyType struct{}

var traceIDKey traceIDKeyType

// Trace decorates a request's context with a trace ID.
// The ID for the request is obtained from the X-Trace-Id field in the request header.
// If that field does not exist or is empty,
// Idempotency-Key and X-Idempotency-Key are tried.
// Failing those, a randomly generated ID is used.
//
// The trace ID can be retrieved from a context so decorated using [TraceID].
// Any trace ID present will be included in log lines generated by [Log].
func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		traceID, err := getTraceID(req)
		if err != nil {
			Errf(w, http.StatusInternalServerError, "getting trace ID: %s", err)
			return
		}

		ctx := req.Context()
		ctx = context.WithValue(ctx, traceIDKey, traceID)
		req = req.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

func getTraceID(req *http.Request) (string, error) {
	for _, field := range []string{"X-Trace-Id", "Idempotency-Key", "X-Idempotency-Key"} {
		traceID := req.Header.Get(field)
		traceID = strings.TrimSpace(traceID)
		if traceID != "" {
			return traceID, nil
		}
	}

	var buf [16]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return "", errors.Wrap(err, "computing random trace ID")
	}

	return hex.EncodeToString(buf[:]), nil
}

// TraceID returns the trace ID decorating the given context, if any.
// See [Trace].
func TraceID(ctx context.Context) string {
	val := ctx.Value(traceIDKey)
	str, _ := val.(string)
	return str
}
