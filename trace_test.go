package mid

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrace(t *testing.T) {
	var got string
	h := Trace(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		got = TraceID(ctx)
	}))

	for _, xTraceID := range []string{"", "a"} {
		for _, idempotencyKey := range []string{"", "b"} {
			for _, xIdempotencyKey := range []string{"", "c"} {
				var want string
				switch {
				case xTraceID != "":
					want = xTraceID
				case idempotencyKey != "":
					want = idempotencyKey
				case xIdempotencyKey != "":
					want = xIdempotencyKey
				}

				req := httptest.NewRequest("GET", "/foo", nil)
				if idempotencyKey != "" {
					req.Header.Set("Idempotency-Key", idempotencyKey)
				}
				if xIdempotencyKey != "" {
					req.Header.Set("X-Idempotency-Key", xIdempotencyKey)
				}
				if xTraceID != "" {
					req.Header.Set("X-Trace-Id", xTraceID)
				}

				got = ""
				h.ServeHTTP(nil, req)

				if want == "" {
					if got == "" {
						t.Error("got empty trace ID, want random")
					}
				} else if got != want {
					t.Errorf("got %s, want %s", got, want)
				}
			}
		}
	}
}
