package middleware_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/kit/http/middleware"
)

func TestApply(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("handler_write\n"))
	})
	var m1 middleware.Middleware = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("middleware1_write\n"))
			h.ServeHTTP(w, r)
		})
	}
	var m2 middleware.Middleware = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("middleware2_write\n"))
			h.ServeHTTP(w, r)
		})
	}
	a := middleware.InFrontOf(h, m1, m2)
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	expected := `middleware1_write
middleware2_write
handler_write
`
	body, err := ioutil.ReadAll(rec.Result().Body)
	require.NoError(t, err)
	assert.Equal(t, expected, string(body))
}
