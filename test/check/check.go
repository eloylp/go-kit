package check

import (
	"crypto/md5" //nolint: gosec
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Function represents a checker function that will
// endure certain results in handler response.
type Function func(t *testing.T, w *http.Response, body []byte)

// Contains check for a valid substring in the response body
// of an HTTP handler.
func Contains(want string) Function {
	return func(t *testing.T, _ *http.Response, body []byte) {
		require.Contains(t, want, string(body))
	}
}

// MatchesMD5 checks that the response body of an HTTP handler
// is equal to the wanted MD5 sum.
func MatchesMD5(want string) Function {
	return func(t *testing.T, _ *http.Response, body []byte) {
		m := md5.New() // nolint: gosec
		if _, err := m.Write(body); err != nil {
			t.Fatal(fmt.Errorf("matchMD5: internal error, cannot write to data stream"))
		}
		got := fmt.Sprintf("%x", m.Sum(nil))
		require.Equal(t, want, got)
	}
}

// HasStatus checks for the specified HTTP status code in the
// handler response.
func HasStatus(want int) Function {
	return func(t *testing.T, w *http.Response, body []byte) {
		require.Equal(t, want, w.StatusCode)
	}
}

// HasHeaders checks for wanted headers and will return an error
// with the non matched ones.
// This only checks that the specified headers are present with its values.
// This means that it could be more headers than the expected. In this case,
// the test will pass.
func HasHeaders(want http.Header) Function {
	return func(t *testing.T, w *http.Response, body []byte) {
		require.Equal(t, want, w.Header)
	}
}

// ContainsJSON checks for equality in for a JSON given
// against the response body of an HTTP handler.
func ContainsJSON(want string) Function {
	return func(t *testing.T, _ *http.Response, body []byte) {
		require.JSONEq(t, want, string(body), "body not contains expected json")
	}
}
