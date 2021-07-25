# Mid - Assorted middleware for HTTP services.

[![Go Reference](https://pkg.go.dev/badge/github.com/bobg/mid.svg)](https://pkg.go.dev/github.com/bobg/mid)
[![Go Report Card](https://goreportcard.com/badge/github.com/bobg/mid)](https://goreportcard.com/report/github.com/bobg/mid)
![Tests](https://github.com/bobg/mid/actions/workflows/go.yml/badge.svg)

Included in this package:

- `Err`, for wrapping handler functions that return Go errors
- `JSON`, for wrapping handler functions that accept and return JSON data
- `Trace`, for tracing a request through its server-side lifetime
- `Log`, for wrapping handler functions to write log messages on entry and exit
- `CodeErr`, an error type for reporting HTTP status codes

plus supporting types and logic.
