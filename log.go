package mid

import (
	"log"
	"net/http"
)

// Log adds logging on entry to and exit from an http.Handler.
//
// If the request is decorated with a trace ID
// (see Trace),
// it is included in the generated log lines.
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		traceID := TraceID(ctx)

		if traceID != "" {
			log.Printf("< %s %s [%s]", req.Method, req.URL, traceID)
		} else {
			log.Printf("< %s %s", req.Method, req.URL)
		}

		ww := ResponseWrapper{W: w}
		next.ServeHTTP(&ww, req)

		if traceID != "" {
			log.Printf("> %d %s %s [%s]", ww.Code, req.Method, req.URL, traceID)
		} else {
			log.Printf("> %d %s %s", ww.Code, req.Method, req.URL)
		}
	})
}
