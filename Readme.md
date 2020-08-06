# Mid - Assorted middleware for HTTP services.

[![GoDoc](https://godoc.org/github.com/bobg/mid?status.svg)](https://godoc.org/github.com/bobg/mid)
[![Go Report Card](https://goreportcard.com/badge/github.com/bobg/mid)](https://goreportcard.com/report/github.com/bobg/mid)

Included in this package:

- `Err`, for wrapping handler functions that return Go errors
- `JSON`, for wrapping handler functions that accept and return JSON data
- `Log`, for wrapping handler functions to write log messages on entry and exit
- `CodeErr`, an error type for reporting HTTP status codes

plus supporting types and logic.
