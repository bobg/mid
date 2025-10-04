package mid

import (
	"context"
	"errors"
)

// Service is an abstraction for any service that can process requests and return responses.
type Service[Req, Resp any] interface {
	Do(context.Context, Req) (Resp, error)
}

// Cache is an abstraction for a cache mapping keys of type Key to values of type Val.
// Implementations must be safe for concurrent use by multiple goroutines.
type Cache[Key, Val any] interface {
	// Get must return [ErrNotFound] if the key is not present in the cache.
	Get(context.Context, Key) (Val, error)

	// Put stores the given value in the cache under the given key.
	Put(context.Context, Key, Val) error
}

// ErrNotFound is the error that [Cache.Get] must return on a cache miss.
var ErrNotFound = errors.New("not found")

// CachingService is a [Service] that wraps another Service.
// On each request, it checks the cache for a response and returns that if found.
// Otherwise it delegates to the wrapped Service.
// On success, that response is stored in the cache before being returned.
type CachingService[Req, Resp any] struct {
	C    Cache[Req, Resp]
	Next Service[Req, Resp]
}

// Do implements the Service interface.
// It first checks the cache for a response for the given request.
// If found, it returns that response.
// Otherwise it delegates to the wrapped Service.
// On success, it stores the response in the cache before returning it.
//
// If the delegated call succeeds but storing the response in the cache fails,
// Do returns the response and a [CachePutError] containing the error from the cache.
func (c CachingService[Req, Resp]) Do(ctx context.Context, req Req) (Resp, error) {
	resp, err := c.C.Get(ctx, req)
	if err == nil {
		return resp, nil
	}
	if !errors.Is(err, ErrNotFound) {
		var zero Resp
		return zero, err
	}

	resp, err = c.Next.Do(ctx, req)
	if err != nil {
		return resp, err
	}
	if err := c.C.Put(ctx, req, resp); err != nil {
		return resp, CachePutError{Err: err}
	}
	return resp, nil
}

// CachePutError is the error type returned by [CachingService.Do]
// when storing a response in the cache fails.
type CachePutError struct {
	Err error
}

// Error implements the error interface.
func (e CachePutError) Error() string {
	return "cache put error: " + e.Err.Error()
}

// Unwrap returns the underlying error.
func (e CachePutError) Unwrap() error {
	return e.Err
}

// LimitedService is a [Service] that wraps another Service.
// Before delegating to the wrapped Service, it waits on the [Limiter] in L.
type LimitedService[Req, Resp any] struct {
	L    Limiter
	Next Service[Req, Resp]
}

// Do implements the Service interface.
func (l LimitedService[Req, Resp]) Do(ctx context.Context, req Req) (Resp, error) {
	if err := l.L.Wait(ctx); err != nil {
		var zero Resp
		return zero, err
	}
	return l.Next.Do(ctx, req)
}

// CacheAndLimitService produces a [Service] by wrapping another Service with caching and rate-limiting layers.
func CacheAndLimitService[Req, Resp any](s Service[Req, Resp], c Cache[Req, Resp], l Limiter) Service[Req, Resp] {
	return CachingService[Req, Resp]{
		C: c,
		Next: LimitedService[Req, Resp]{
			L:    l,
			Next: s,
		},
	}
}
