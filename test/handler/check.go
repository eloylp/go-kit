package handler

import (
	"crypto/md5" //nolint: gosec
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// CheckerFunc represents a checker function that will
// endure certain results in handler response.
type CheckerFunc func(t *testing.T, w *http.Response, body []byte)

// CheckContains check for a valid substring in the response body
// of an HTTP handler.
func CheckContains(want string) CheckerFunc {
	return func(t *testing.T, _ *http.Response, body []byte) {
		require.Contains(t, string(body), want)
	}
}

// CheckMatchesMD5 checks that the response body of an HTTP handler
// is equal to the wanted MD5 sum.
func CheckMatchesMD5(want string) CheckerFunc {
	return func(t *testing.T, _ *http.Response, body []byte) {
		m := md5.New() // nolint: gosec
		if _, err := m.Write(body); err != nil {
			t.Fatal(fmt.Errorf("matchMD5: internal error, cannot write to data stream"))
		}
		got := fmt.Sprintf("%x", m.Sum(nil))
		require.Equal(t, want, got)
	}
}

// CheckHasStatus checks for the specified HTTP status code in the
// handler response.
func CheckHasStatus(want int) CheckerFunc {
	return func(t *testing.T, w *http.Response, body []byte) {
		require.Equal(t, want, w.StatusCode)
	}
}

// CheckHasHeaders checks for wanted headers and will return an error
// with the non matched ones.
// This only checks that the specified headers are present with its values.
// This means that it could be more headers than the expected. In this case,
// the test will pass.
func CheckHasHeaders(want http.Header) CheckerFunc {
	return func(t *testing.T, w *http.Response, body []byte) {
		require.Equal(t, want, w.Header)
	}
}

// CheckContainsJSON checks for equality in for a JSON given
// against the response body of an HTTP handler.
func CheckContainsJSON(want string) CheckerFunc {
	return func(t *testing.T, _ *http.Response, body []byte) {
		require.JSONEq(t, want, string(body), "body not contains expected json")
	}
}
