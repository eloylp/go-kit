package middleware_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/kit/http/middleware"
)

type EndpointMapperMock struct{}

func (e EndpointMapperMock) Map(url string) string {
	return "mapped-" + url
}

func TestRequestDurationObserver(t *testing.T) {
	// Prepare prometheus registry
	reg := prometheus.NewRegistry()
	mid := middleware.RequestDurationObserver("app", reg, []float64{0.05, 0.08, 0.1, 0.2}, EndpointMapperMock{})

	mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte("hello"))
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/product", nil))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ph := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	ph.ServeHTTP(rec, req)

	respMetrics, err := ioutil.ReadAll(rec.Body)
	require.NoError(t, err)
	metrics := string(respMetrics)

	assert.Contains(t, metrics, `# TYPE app_http_request_duration_seconds histogram`)
	assert.Contains(t, metrics, `app_http_request_duration_seconds_bucket{code="200",endpoint="mapped-/product",method="GET",le="0.05"} 0`)
	assert.Contains(t, metrics, `app_http_request_duration_seconds_bucket{code="200",endpoint="mapped-/product",method="GET",le="0.08"} 0`)
	assert.Contains(t, metrics, `app_http_request_duration_seconds_bucket{code="200",endpoint="mapped-/product",method="GET",le="0.1"} 0`)
	assert.Contains(t, metrics, `app_http_request_duration_seconds_bucket{code="200",endpoint="mapped-/product",method="GET",le="0.2"} 1`)
	assert.Contains(t, metrics, `app_http_request_duration_seconds_bucket{code="200",endpoint="mapped-/product",method="GET",le="+Inf"} 1`)
	assert.Contains(t, metrics, `app_http_request_duration_seconds_sum{code="200",endpoint="mapped-/product",method="GET"}`)
	assert.Contains(t, metrics, `app_http_request_duration_seconds_count{code="200",endpoint="mapped-/product",method="GET"} 1`)
}
