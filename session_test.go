package mid

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testSession struct {
	id      int
	csrfKey [sha256.Size]byte
}

func (s testSession) CSRFKey() [sha256.Size]byte { return s.csrfKey }
func (testSession) Active() bool                 { return true }
func (testSession) Exp() time.Time               { return time.Now().Add(24 * time.Hour) }

func TestCSRF(t *testing.T) {
	var s1, s2 testSession
	s2.csrfKey[0] = 'x'

	tok, err := CSRFToken(s1)
	if err != nil {
		t.Fatal(err)
	}

	err = CSRFCheck(s1, tok)
	if err != nil {
		t.Errorf("got error %s, want nil", err)
	}
	err = CSRFCheck(s2, tok)
	if !errors.Is(err, ErrCSRF) {
		t.Errorf("got error %v, want %s", err, ErrCSRF)
	}
	err = CSRFCheck(s1, tok[1:])
	if !errors.Is(err, ErrCSRF) {
		t.Errorf("got error %v, want %s", err, ErrCSRF)
	}
	err = CSRFCheck(s1, "foo")
	if !errors.Is(err, ErrCSRF) {
		t.Errorf("got error %v, want %s", err, ErrCSRF)
	}
}

func TestSessionHandler(t *testing.T) {
	var (
		store testSessionStore
		got   int
	)
	server := httptest.NewServer(SessionHandler(store, "cookie", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sess := ContextSession(req.Context())
		if sess == nil {
			got = 0
		} else {
			got = sess.(testSession).id
		}
	})))
	defer server.Close()

	cases := []struct {
		val              string
		want, wantStatus int
	}{{
		val:        "foo",
		wantStatus: http.StatusNoContent,
		want:       1,
	}, {
		val:        "bar",
		wantStatus: http.StatusForbidden,
	}, {
		wantStatus: http.StatusForbidden,
	}}

	var client http.Client

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case_%02d", i+1), func(t *testing.T) {
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			if tc.val != "" {
				req.AddCookie(&http.Cookie{Name: "cookie", Value: tc.val})
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tc.wantStatus)
			}

			if resp.StatusCode == http.StatusNoContent && got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

type testSessionStore struct{}

func (s testSessionStore) GetSession(_ context.Context, key string) (Session, error) {
	if key == "foo" {
		return testSession{id: 1}, nil
	}
	return nil, ErrNoSession
}

func (testSessionStore) CancelSession(context.Context, string) error {
	return fmt.Errorf("unimplemented")
}
