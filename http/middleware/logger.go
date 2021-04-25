package middleware

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// RequestLogger will log the client request
// information on each request at Debug level.
// Accepts logrus as logger.
func RequestLogger(logger *logrus.Logger) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.WithFields(logrus.Fields{
				"path":    r.URL.String(),
				"method":  r.Method,
				"ip":      r.RemoteAddr,
				"headers": r.Header,
			}).Debug("intercepted request")
			h.ServeHTTP(w, r)
		})
	}
}
