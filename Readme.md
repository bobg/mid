# Mid - Assorted middleware for HTTP services.

[![Go Reference](https://pkg.go.dev/badge/github.com/bobg/mid.svg)](https://pkg.go.dev/github.com/bobg/mid)
[![Go Report Card](https://goreportcard.com/badge/github.com/bobg/mid)](https://goreportcard.com/report/github.com/bobg/mid)
![Tests](https://github.com/bobg/mid/actions/workflows/go.yml/badge.svg)

This is mid,
a collection of useful middleware for HTTP services.

## Err

The `Err` function turns a `func(http.ResponseWriter, *http.Request) error`
into an `http.Handler`,
allowing handler functions to return errors in a more natural fashion.
These errors result in the handler producing an HTTP 500 status code,
but `CodeErr` and `Responder` allow you to control this behavior
(see below).
A `nil` return produces a `200`,
or a `204` (“no content”) if no bytes were written to the response object.

Usage:

```go
func main() {
  http.Handle("/foo", Err(fooHandler))
  http.ListenAndServe(":8080", nil)
}

func fooHandler(w http.ResponseWriter, req *http.Request) error {
  if x := req.FormValue("x"); x != "secret password" {
    return CodeErr{C: http.StatusUnauthorized}
  }
  fmt.Fprintf(w, "You know the secret password")
  return nil
}
```

## JSON

The `JSON` function turns a `func(context.Context, X) (Y, error)`
into an `http.Handler`,
where `X` is the type of a parameter into which the HTTP request body is automatically JSON-unmarshaled,
and `Y` is the type of a result that is automatically JSON-marshaled into the HTTP response.
Any error is handled as with `Err`
(see above).
All parts of the `func` signature are optional:
the `context.Context` parameter,
the `X` parameter,
the `Y` result,
and the `error` result.

Usage:

```go
func main() {
  // Parses the request body as a JSON-encoded array of strings,
  // then sorts, re-encodes, and returns that array.
  http.Handle("/bar", JSON(barHandler))

  http.ListenAndServe(":8080", nil)
}

func barHandler(inp []string) []string {
  sort.Strings(inp)
  return inp
}
```

## CodeErr and Responder

`CodeErr` is an `error` type suitable for returning from `Err`- and `JSON`-wrapped handlers that can control the HTTP status code that gets returned.
It contains an HTTP status code field and a nested `error`.

`Responder` is an interface
(implemented by `CodeErr`)
that allows an error type to control how `Err`- and `JSON`-wrapped handlers respond to the pending request.

## ResponseWrapper

`ResponseWrapper` is an `http.ResponseWriter` that wraps a nested `http.ResponseWriter` and also records the status code and number of bytes sent in the response.

## Trace

The `Trace` function wraps an `http.Handler` and decorates the `context.Context` in its `*http.Request` with any “trace ID” string found in the request header.

## Log

The `Log` function wraps an `http.Handler` with a function that writes a simple log line on the way into and out of the handler.
The log line includes any “trace ID” found in the request’s `context.Context`.
