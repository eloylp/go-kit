package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.eloylp.dev/kit/test/check"
)

// Case represents a test case for an HTTP handler.
// The field Case must be a brief description of what
// is being tested.
// Checkers field is an slice of check.Function, that will
// be run against the response of the router.
// The rest of the fields are what forms the HTTP request
// for running the handler under test.
type Case struct {
	Case, Path, Method string
	Body               io.Reader
	Headers            http.Header
	Checkers           []check.Function
}

// TestAuxFunc its the type for functions used to
// bring up/down services, as databases, before and after
// a handler test is executed.
type TestAuxFunc func(t *testing.T)

// Tester will take all the handler test cases and run them serialized.
// It will execute the corresponding setUp and tearDown TestAuxFunc functions,
// passed as parameters, that need to include the setup and teardown logic. Probably,
// this functions will need to implement t.Fatal() to not continue executing current test.
// For the TestAuxFunc you can still pass nil , if yo dont have any logic for them.
func Tester(cases []Case, router http.Handler, setUp, tearDown TestAuxFunc) func(t *testing.T) {
	return func(t *testing.T) {
		for _, tt := range cases {
			name := fmt.Sprintf("path: %s, method: %s, case: %q", tt.Path, tt.Method, tt.Case)
			t.Run(name, func(t *testing.T) {
				if setUp != nil {
					setUp(t)
				}
				rec, req := httptest.NewRecorder(), httptest.NewRequest(tt.Method, tt.Path, tt.Body) //nolint:scopelint
				req.Header = tt.Headers                                                              //nolint:scopelint
				router.ServeHTTP(rec, req)
				res := rec.Result() //nolint:bodyclose
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				for _, chk := range tt.Checkers { //nolint:scopelint
					if err := chk(res, body); err != nil {
						t.Errorf("FAILED %s: %v", name, err)
					}
				}
				if tearDown != nil {
					tearDown(t)
				}
			})
		}
	}
}
