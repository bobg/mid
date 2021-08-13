package mid

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestLogTrace(t *testing.T) {
	s := httptest.NewServer(Trace(Log(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))))
	defer s.Close()

	var (
		c           http.Client
		wantLine1RE = regexp.MustCompile(`< GET /foo \[(.*)\]`)
		wantLine2RE = regexp.MustCompile(`> 0 GET /foo \[(.*)\]`)
	)

	const traceID = "xyzzy"

	fields := []string{"", "X-Trace-Id", "Idempotency-Key", "X-Idempotency-Key"}
	for _, field := range fields {
		t.Run(fmt.Sprintf("case_%s", field), func(t *testing.T) {
			got := new(bytes.Buffer)
			log.SetOutput(got)

			if field == "" {
				_, err := http.Get(s.URL + "/foo")
				if err != nil {
					t.Fatal(err)
				}
			} else {
				req, err := http.NewRequest("GET", s.URL+"/foo", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set(field, traceID)
				_, err = c.Do(req)
				if err != nil {
					t.Fatal(err)
				}
			}

			sc := bufio.NewScanner(got)
			if !sc.Scan() {
				t.Fatal("could not read line 1 of output")
			}
			line1 := sc.Text()
			if !sc.Scan() {
				t.Fatal("could not read line 2 of output")
			}
			line2 := sc.Text()
			if err := sc.Err(); err != nil {
				t.Fatal(err)
			}

			m := wantLine1RE.FindStringSubmatch(line1)
			if len(m) != 2 {
				t.Errorf("line 1 mismatch: %s", line1)
			} else if field == "" {
				trimmed := strings.Trim(m[1], "0123456789abcdef")
				if trimmed != "" {
					t.Errorf("line 1 trace ID mismatch: %s", m[1])
				}
			} else if m[1] != traceID {
				t.Errorf("line 1 trace ID mismatch: %s", m[1])
			}

			m = wantLine2RE.FindStringSubmatch(line2)
			if len(m) != 2 {
				t.Errorf("line 2 mismatch: %s", line2)
			} else if field == "" {
				trimmed := strings.Trim(m[1], "0123456789abcdef")
				if trimmed != "" {
					t.Errorf("line 2 trace ID mismatch: %s", m[1])
				}
			} else if m[1] != traceID {
				t.Errorf("line 2 trace ID mismatch: %s", m[1])
			}
		})
	}
}
