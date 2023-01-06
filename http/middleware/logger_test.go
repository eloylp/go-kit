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

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.RemoteAddr = "192.168.1.15"
	req.Header.Add("Accept", "application/json")
	mid := middleware.RequestLogger(logger)
	mid(nullHandler).ServeHTTP(rec, req)

	logs := logOut.String()
	assert.Contains(t, logs, "method=GET")
	assert.Contains(t, logs, "path=/path")
	assert.Contains(t, logs, "ip=192.168.1.15")
	assert.Contains(t, logs, "headers=\"map[Accept:[application/json]]")
	assert.Contains(t, logs, "duration")
	assert.Contains(t, logs, "msg=\"intercepted request\"")
	assert.Contains(t, logs, "level=debug")
}
