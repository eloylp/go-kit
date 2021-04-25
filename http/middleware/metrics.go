package middleware

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// RequestDurationObserver will observe the duration of all the requests
// and register them on a given Prometheus registry. It will use an histogram
// where the client code can define its custom buckets. The client application
// could also specify a namespace to not collide with other similar metrics names.
func RequestDurationObserver(namespace string, registry prometheus.Registerer, buckets []float64) Middleware {
	observer := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "http_request",
		Name:      "duration_seconds",
		Buckets:   buckets,
	})
	registry.MustRegister(observer)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			h.ServeHTTP(w, r)
			duration := time.Since(start)
			observer.Observe(duration.Seconds())
		})
	}
}
