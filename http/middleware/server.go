package middleware

import (
	"net/http"
)

// ServerHeader will grab server information in the
// "Server" header for all the requests.
func ServerHeader(value string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Server", value)
			h.ServeHTTP(w, r)
		})
	}
}
