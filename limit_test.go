package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestLimiter(t *testing.T) {
	cases := []struct {
		waitErr, rtErr error
	}{{
		waitErr: nil,
		rtErr:   nil,
	}, {
		waitErr: errors.New("wait error"),
	}, {
		rtErr: errors.New("round trip error"),
	}}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case_%02d", i+1), func(t *testing.T) {
			var (
				lim    = mockLimiter{waitErr: tc.waitErr}
				transp = mockTransport{rtErr: tc.rtErr}
				lt     = LimitedTransport{L: lim, T: transp}
			)

			_, err := lt.RoundTrip(&http.Request{})

			switch {
			case tc.waitErr != nil:
				if !errors.Is(err, tc.waitErr) {
					t.Errorf("got %v, want %v", err, tc.waitErr)
				}

			case tc.rtErr != nil:
				if !errors.Is(err, tc.rtErr) {
					t.Errorf("got %v, want %v", err, tc.rtErr)
				}

			default:
				if err != nil {
					t.Errorf("got %v, want nil", err)
				}
			}
		})
	}
}

type mockLimiter struct {
	waitErr error
}

func (ml mockLimiter) Wait(context.Context) error {
	return ml.waitErr
}

type mockTransport struct {
	rtErr error
}

func (mt mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, mt.rtErr
}
