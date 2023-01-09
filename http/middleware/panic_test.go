package middleware_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.eloylp.dev/kit/http/middleware"
)

func TestPanic(t *testing.T) {

	logBuff := bytes.NewBuffer(nil)

	logger := log.New(logBuff, "test-", log.Default().Flags())

	handlerFunc := func(v interface{}) {
		logger.Printf("panic detected: %v\n", v)
	}

	mid := middleware.PanicHandler(handlerFunc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		panic("panic !")

	})).ServeHTTP(rec, req)

	assert.Contains(t, logBuff.String(), "panic detected: panic !")
}
