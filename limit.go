package mid

import (
	"context"
	"net/http"
)

// Limiter is the type of an object that can be used to limit the rate of some operation.
// Calling Wait on a Limiter blocks until the operation is allowed to proceed.
//
// This interface is satisfied by the *Limiter type in golang.org/x/time/rate.
type Limiter interface {
	Wait(context.Context) error
}

// LimitedTransport is an [http.RoundTripper] that limits the rate of requests it makes using a [Limiter].
// After waiting for the limiter in L, it delegates to the http.RoundTripper in T.
// If T is nil, it uses [http.DefaultTransport].
type LimitedTransport struct {
	L Limiter
	T http.RoundTripper
}

// RoundTrip implements the [http.RoundTripper] interface.
func (lt LimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := lt.L.Wait(req.Context()); err != nil {
		return nil, err
	}
	next := lt.T
	if next == nil {
		next = http.DefaultTransport
	}
	return next.RoundTrip(req)
}
