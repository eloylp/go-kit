package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// RequestLogger will log the client request
// information on each request at Debug level.
// Accepts logrus as logger.
func RequestLogger(logger *logrus.Entry, level logrus.Level) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.WithFields(logrus.Fields{
				"path":     r.URL.String(),
				"method":   r.Method,
				"ip":       r.RemoteAddr,
				"headers":  r.Header,
				"duration": duration,
			}).Log(level, "intercepted request")
		})
	}
}
