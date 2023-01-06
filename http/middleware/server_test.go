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
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mid := middleware.ServerHeader("the server")
	mid(nullHandler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	assert.Contains(t, "the server", rec.Result().Header.Get("Server"))
}
