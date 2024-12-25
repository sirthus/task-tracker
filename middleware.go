package main

import (
	"log"
	"net/http"
	"time"
)

// LogRequestDuration logs the method, path, and duration of each request
func LogRequestDuration(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// capture current time for logging duration
		start := time.Now()

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		duration := time.Since(start)

		log.Printf("Handled %s %s in %v", r.Method, r.URL.Path, duration)

	})

}

// ValidateJSON ensures the request Content-Type is application/json
func ValidateJSON(next http.Handler, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method matches any specified method
		for _, method := range methods {
			if r.Method == method {
				// Check if the Content-Type header is not application/json
				if r.Header.Get("Content-Type") != "application/json" {
					http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
					return
				}
				break
			}
		}
		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
