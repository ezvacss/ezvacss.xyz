package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger is a middleware that logs all incoming requests.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Use a custom ResponseWriter to capture the status code if needed
		// Here we just log the request
		next.ServeHTTP(w, r)
		
		log.Printf("%s %s %s %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}
