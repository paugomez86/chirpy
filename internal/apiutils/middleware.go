package apiutils

import (
	"log"
	"net/http"
)

// Middleware to increment the hits to the file server
func (api *ApiUtils) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Middleware wrapper function to log the request parameters
func MiddlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
