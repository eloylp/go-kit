package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Case represents a test case for an HTTP handler.
type Case struct {
	// Case is name for the given test scenario.
	Case,
	// Following fields represents the data for the request
	Path,
	Method string
	Body    io.Reader
	Headers http.Header
	// Checkers is list of checkers to apply to the response of the request.
	// See functions like handler.CheckContains() on this package.
	Checkers []CheckerFunc
	// The setUp functions optionally allows configuring test dependencies,
	// like databases.
	setUp TestAuxFunc
	// The tearDown functions optionally allows shutdown test dependencies,
	// like databases.
	tearDown TestAuxFunc
}

// TestAuxFunc its the type for functions used to
// bring up/down services, as databases, before and after
// a handler test is executed.
type TestAuxFunc func(t *testing.T)

// Tester will test the provided HTTP handler by executing
// take all the test []Case in a serialized way.
//
// If provided, it will execute the corresponding
// setUp and tearDown TestAuxFunc functions per each
// test case. See Case type on how to configure them.
func Tester(t *testing.T, cases []Case, handler http.Handler) {

	for _, tt := range cases {
		if tt.setUp != nil {
			tt.setUp(t)
		}
		name := fmt.Sprintf("path: %s, method: %s, case: %q", tt.Path, tt.Method, tt.Case)
		t.Run(name, func(t *testing.T) {
			rec, req := httptest.NewRecorder(), httptest.NewRequest(tt.Method, tt.Path, tt.Body) //nolint:scopelint
			req.Header = tt.Headers                                                              //nolint:scopelint
			handler.ServeHTTP(rec, req)
			res := rec.Result()
			// Save a copy of the body before checkers execution ...
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			for _, chk := range tt.Checkers { //nolint:scopelint
				// The body is a buffer that only can be consumed once. Send a new reader every time,
				// so checkers don't need to worry about restoring the body.
				res.Body = io.NopCloser(bytes.NewReader(body))
				chk(t, res)
			}
		})
		if tt.tearDown != nil {
			tt.tearDown(t)
		}
	}
}
