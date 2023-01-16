package handler

import (
	"crypto/md5" //nolint: gosec
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// CheckerFunc represents a checker that will
// assert certain data in handler response.
type CheckerFunc func(t *testing.T, resp *http.Response)

// CheckContains checks for a valid substring in the response body
// of an HTTP handler.
func CheckContains(want string) CheckerFunc {
	return func(t *testing.T, resp *http.Response) {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		require.Contains(t, string(data), want)
	}
}

// CheckMatchesMD5 checks that the response body of an HTTP handler
// is equal to the expected MD5 sum.
func CheckMatchesMD5(expectedMD5 string) CheckerFunc {
	return func(t *testing.T, resp *http.Response) {
		m := md5.New() // nolint: gosec

		if _, err := io.Copy(m, resp.Body); err != nil {
			t.Fatal(fmt.Errorf("matchMD5: internal error, cannot write to data stream"))
		}
		got := fmt.Sprintf("%x", m.Sum(nil))
		require.Equal(t, expectedMD5, got)
	}
}

// CheckHasStatus checks for the specified HTTP status code in the
// handler response.
func CheckHasStatus(want int) CheckerFunc {
	return func(t *testing.T, w *http.Response) {
		require.Equal(t, want, w.StatusCode)
	}
}

// CheckHasHeaders checks the wanted headers are present in the response.
//
// This only checks that the specified headers are present with its
// respective values. This means that it could be more headers than the expected.
// In this case the test will pass.
func CheckHasHeaders(want http.Header) CheckerFunc {
	return func(t *testing.T, w *http.Response) {
		require.Equal(t, want, w.Header)
	}
}

// CheckContainsJSON checks that the provided JSON string
// matches the one in the response body.
func CheckContainsJSON(expectedJSON string) CheckerFunc {
	return func(t *testing.T, resp *http.Response) {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		require.JSONEq(t, expectedJSON, string(data), "body not contains expected json")
	}
}
