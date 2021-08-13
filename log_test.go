package mid

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogTrace(t *testing.T) {
	s := httptest.NewServer(Trace(Log(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))))
	defer s.Close()

	var c http.Client

	got := new(bytes.Buffer)
	log.SetOutput(got)

	req, err := http.NewRequest("GET", s.URL+"/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Trace-Id", "xyzzy")
	_, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	sc := bufio.NewScanner(got)
	if !sc.Scan() {
		t.Fatal("could not read line 1 of output")
	}
	line1 := sc.Text()
	if !sc.Scan() {
		t.Fatal("could not read line 2 of output")
	}
	if !strings.Contains(line1, "< GET /foo [xyzzy]") {
		t.Errorf("line 1 mismatch: %s", line1)
	}

	line2 := sc.Text()
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(line2, "> 204 GET /foo [xyzzy]") {
		t.Errorf("line 2 mismatch: %s", line2)
	}
}
