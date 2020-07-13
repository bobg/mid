package mid

import (
	"log"
	"net/http"
)

// Log adds logging on entry to and exit from an http.Handler.
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("< %s %s", req.Method, req.URL)
		ww := responseWriterWrapper{w: w}
		next.ServeHTTP(&ww, req)
		log.Printf("> %d %s %s", ww.code, req.Method, req.URL)
	})
}
