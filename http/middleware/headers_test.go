//go:build unit

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/kit/http/middleware"
)

func TestServerHeader(t *testing.T) {

	defaultHeaders := middleware.DefaultHeaders{}
	defaultHeaders.Set("Server", "a server")

	mid := middleware.ResponseHeaders(defaultHeaders)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
	})).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	assert.Equal(t, "a server", rec.Header().Get("Server"), "it should rewrite the specified header")
	assert.Equal(t, "custom-value", rec.Header().Get("X-Custom-Header"), "it should not rewrite other headers")
}
