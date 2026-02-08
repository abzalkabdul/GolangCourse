package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const validAPIKey = "secret12345"

func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")

		if apiKey != validAPIKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamp := time.Now().Format("2006-01-02T15:04:05")
		log.Printf("%s %s %s Request received", timestamp, r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
