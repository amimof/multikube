package multikube

import (
	"net/http"
	"log"
)

type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc

func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Context: %s", r.Header.Get("X-Multikube-Context"))
		next.ServeHTTP(w, r)
	}
}

func WithContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("X-Multikube-Context", "test")
		next.ServeHTTP(w, r)
	}
}