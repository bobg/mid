package mid

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestErrf(t *testing.T) {
	cases := []struct {
		code     int
		format   string
		wantCode int
		wantStr  string
	}{{
		code: 0, format: "", wantCode: http.StatusInternalServerError, wantStr: http.StatusText(http.StatusInternalServerError),
	}, {
		code: 0, format: "foo", wantCode: http.StatusInternalServerError, wantStr: "foo",
	}}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%02d", i+1), func(t *testing.T) {
			var got testWriter
			Errf(&got, c.code, c.format)
			if got.code != c.wantCode {
				t.Errorf("got code %d, want %d", got.code, c.wantCode)
			}
			if strings.TrimSpace(got.str) != c.wantStr {
				t.Errorf(`got string "%s", want "%s"`, got.str, c.wantStr)
			}
		})
	}
}

type testWriter struct {
	code int
	str  string
}

func (w *testWriter) Header() http.Header {
	return http.Header{}
}

func (w *testWriter) Write(b []byte) (int, error) {
	w.str += string(b)
	return len(b), nil
}

func (w *testWriter) WriteHeader(code int) {
	w.code = code
}
