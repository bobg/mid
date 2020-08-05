package mid

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErr(t *testing.T) {
	var (
		e1 = errors.New("e1")
		e2 = CodeErr{C: http.StatusMethodNotAllowed}
	)

	s := httptest.NewServer(Err(func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/b":
			return e1

		case "/c":
			return e2

		case "/d":
			w.Write([]byte("foo"))
		}
		return nil
	}))

	cases := []struct {
		path     string
		wantCode int
		wantBody string
	}{
		{
			path:     "/a",
			wantCode: http.StatusNoContent,
		},
		{
			path:     "/b",
			wantCode: http.StatusInternalServerError,
		},
		{
			path:     "/c",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			path:     "/d",
			wantCode: http.StatusOK,
			wantBody: "foo",
		},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			resp, err := http.Get(s.URL + c.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != c.wantCode {
				t.Errorf("got code %d, want %d", resp.StatusCode, c.wantCode)
			}
			if c.wantBody != "" {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(body) != c.wantBody {
					t.Errorf("got body \"%s\", want \"%s\"", string(body), c.wantBody)
				}
			}
		})
	}
}
