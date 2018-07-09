package server

import "net/http"

func CORSHandler(next http.Handler, corsAllowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", corsAllowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

		// stop here if the request is an OPTIONS preflight
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)

		return
	})
}
