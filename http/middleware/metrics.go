package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/negroni"
)

type EndpointMapper interface {
	Map(url string) string
}

// RequestDurationObserver will observe the duration of all the requests
// and register them on a given Prometheus registry. It will use an histogram
// where the client code can define its custom buckets. The client application
// could also specify a namespace to not collide with other similar metrics names in the same runtime.
func RequestDurationObserver(namespace string, registry prometheus.Registerer, buckets []float64, endpointMapper EndpointMapper) Middleware {
	observer := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "http_request",
		Name:      "duration_seconds",
		Buckets:   buckets,
	}, []string{"method", "code", "endpoint"})
	registry.MustRegister(observer)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nw := negroni.NewResponseWriter(w)
			start := time.Now()
			h.ServeHTTP(nw, r)
			duration := time.Since(start)
			statusCode := strconv.Itoa(nw.Status())
			endpoint := endpointMapper.Map(r.URL.String())
			observer.WithLabelValues(r.Method, statusCode, endpoint).Observe(duration.Seconds())
		})
	}
}
