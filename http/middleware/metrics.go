package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/negroni"
)

// EndpointMapper represents a way to map an http.Request.URL.String(),
// to its respective handler. This responsibility is delegated to the
// client code. http.Request.URL.String() cannot be mapped directly
// to a metrics label due to combinatorial explosion. This means a path
// can contain /product/[0-9] and each ID will generate a new whole
// time series in a metrics system like prometheus.
// The recommendation is to wrap some efficient implementation for
// properly matching prefixes. The recommendation is to use something
// like a Radix tree implementation. See  https://github.com/hashicorp/go-immutable-radix .
type EndpointMapper interface {
	Map(url string) string
}

// RequestDurationObserver will observe the duration of all the requests
// and register them on a given Prometheus registry. It will use an histogram
// where the client code can define its custom buckets. The client application
// could also specify a namespace to not collide with other similar metrics names in the same runtime.
// Labels will reflect the HTTP method, code and the endpoint mapped by EndpointMapper.
func RequestDurationObserver(namespace string, registry prometheus.Registerer, buckets []float64,
	endpointMapper EndpointMapper) Middleware {
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
			endpoint := endpointMapper.Map(r.RequestURI)
			observer.WithLabelValues(r.Method, statusCode, endpoint).Observe(duration.Seconds())
		})
	}
}

// ResponseSizeObserver will observe the size of the body of all the responses
// and register them on a given Prometheus registry. It will use an histogram
// where the client code can define its custom buckets. The client application
// could also specify a namespace to not collide with other similar metrics names in the same runtime.
// Labels will reflect the HTTP method, code and the endpoint mapped by EndpointMapper.
func ResponseSizeObserver(namespace string, registry prometheus.Registerer, buckets []float64, endpointMapper EndpointMapper) Middleware {
	observer := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "http_response",
		Name:      "size",
		Buckets:   buckets,
	}, []string{"method", "code", "endpoint"})
	registry.MustRegister(observer)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nw := negroni.NewResponseWriter(w)
			h.ServeHTTP(nw, r)
			statusCode := strconv.Itoa(nw.Status())
			endpoint := endpointMapper.Map(r.RequestURI)
			observer.WithLabelValues(r.Method, statusCode, endpoint).Observe(float64(nw.Size()))
		})
	}
}
