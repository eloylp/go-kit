package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/eloylp/kit/http/middleware"
)

func TestRequestLogger(t *testing.T) {
	logOut := bytes.NewBuffer(nil)
	logger := logrus.New()
	logger.SetOutput(logOut)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.RemoteAddr = "192.168.1.15"
	req.Header.Add("Accept", "application/json")
	mid := middleware.RequestLogger(logger)
	mid(nullHandler).ServeHTTP(rec, req)

	logs := logOut.String()
	assert.Contains(t, logs, http.MethodGet)
	assert.Contains(t, logs, "/path")
	assert.Contains(t, logs, "192.168.1.15")
	assert.Contains(t, logs, "Accept")
	assert.Contains(t, logs, "application/json")
}
