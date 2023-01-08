//go:build unit

package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/kit/http/middleware"
)

func TestRequestLogger(t *testing.T) {
	logOut := bytes.NewBuffer(nil)

	logger := logrus.New()
	logger.SetOutput(logOut)
	logger.SetLevel(logrus.DebugLevel)

	entry := logrus.NewEntry(logger)

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.RemoteAddr = "192.168.1.15"
	req.Header.Add("Accept", "application/json")

	mid := middleware.RequestLogger(entry, logrus.DebugLevel)

	rec := httptest.NewRecorder()
	mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello !"))
	})).ServeHTTP(rec, req)

	logs := logOut.String()

	assert.Contains(t, logs, "method=GET")
	assert.Contains(t, logs, "path=/path")
	assert.Contains(t, logs, "ip=192.168.1.15")
	assert.Contains(t, logs, "headers=\"map[Accept:[application/json]]")
	assert.Contains(t, logs, "duration")
	assert.Contains(t, logs, "response_size=7")
	assert.Contains(t, logs, "msg=\"intercepted request\"")
	assert.Contains(t, logs, "level=debug")
}
