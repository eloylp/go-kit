//go:build unit

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/kit/http/middleware"
)

// Some pre made data for the tests
var (
	nullHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	userAuth    = middleware.Authorization{
		"user": "$2y$10$mAx10mlJ/UNbQJCgPp2oLe9n9jViYl9vlT0cYI3Nfop3P3bU1PDay", // unencrypted is user:password
	}
)

// Groups all the info needed for execute a test case.
type AuthTestCase struct {
	Name             string
	SutConfig        AuthSutConfig
	TestInput        AuthTestInput
	ExpectedHTTPCode int
}

// AuthSutConfig configures the subject under tests (SUT)
type AuthSutConfig struct {
	Methods     []string
	PathRegexes []string
	Auth        middleware.Authorization
}

// AuthTestInput is the input to be submitted to the SUT.
type AuthTestInput struct {
	User         string
	Password     string
	TargetPath   string
	TargetMethod string
}

//nolint:scopelint
func TestAuthChecker(t *testing.T) {
	cases := authTestCases()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			cfg := middleware.NewAuthConfig().
				WithMethods(c.SutConfig.Methods).
				WithAuth(c.SutConfig.Auth)
			for _, r := range c.SutConfig.PathRegexes {
				cfg.WithPathRegex(r)
			}
			req := httptest.NewRequest(c.TestInput.TargetMethod, c.TestInput.TargetPath, nil)
			if c.TestInput.User != "" {
				req.SetBasicAuth(c.TestInput.User, c.TestInput.Password)
			}
			rec := httptest.NewRecorder()
			mid := middleware.AuthChecker(cfg)
			mid(nullHandler).ServeHTTP(rec, req)
			assert.Equal(t, c.ExpectedHTTPCode, rec.Result().StatusCode)
		})
	}
}

func authTestCases() []AuthTestCase {
	return []AuthTestCase{
		{
			Name: "Authenticated user must access all routes for GET method.",
			SutConfig: AuthSutConfig{
				Methods:     []string{http.MethodGet},
				PathRegexes: []string{"^.*"},
				Auth:        userAuth,
			},
			TestInput: AuthTestInput{
				User:         "user",
				Password:     "password",
				TargetPath:   "/",
				TargetMethod: http.MethodGet,
			},
			ExpectedHTTPCode: http.StatusOK,
		},
		{
			Name: "Authenticated user must access to a sub path.",
			SutConfig: AuthSutConfig{
				Methods:     []string{http.MethodGet},
				PathRegexes: []string{"^.*"},
				Auth:        userAuth,
			},
			TestInput: AuthTestInput{
				User:         "user",
				Password:     "password",
				TargetPath:   "/sub-path",
				TargetMethod: http.MethodGet,
			},
			ExpectedHTTPCode: http.StatusOK,
		},
		{
			Name: "Non authenticated user must not access routes for GET method.",
			SutConfig: AuthSutConfig{
				Methods:     []string{http.MethodGet},
				PathRegexes: []string{"^.*"},
				Auth:        userAuth,
			},
			TestInput: AuthTestInput{
				User:         "user",
				Password:     "non-valid-password",
				TargetPath:   "/",
				TargetMethod: http.MethodGet,
			},
			ExpectedHTTPCode: http.StatusUnauthorized,
		},
		{
			Name: "Non authenticated user must NOT access to a protected sub path.",
			SutConfig: AuthSutConfig{
				Methods:     []string{http.MethodGet},
				PathRegexes: []string{"^.*"},
				Auth:        userAuth,
			},
			TestInput: AuthTestInput{
				User:         "user",
				Password:     "non-valid-password",
				TargetPath:   "/sub-path",
				TargetMethod: http.MethodGet,
			},
			ExpectedHTTPCode: http.StatusUnauthorized,
		},
		{
			Name: "Non authenticated user can access to a non protected sub path.",
			SutConfig: AuthSutConfig{
				Methods:     []string{http.MethodGet},
				PathRegexes: []string{"^/protected.*"},
				Auth:        userAuth,
			},
			TestInput: AuthTestInput{
				TargetPath:   "/non-protected",
				TargetMethod: http.MethodGet,
			},
			ExpectedHTTPCode: http.StatusOK,
		},
		{
			Name: "Non authenticated user can not access to a protected sub path.",
			SutConfig: AuthSutConfig{
				Methods:     []string{http.MethodGet},
				PathRegexes: []string{"^/protected.*"},
				Auth:        userAuth,
			},
			TestInput: AuthTestInput{
				TargetPath:   "/protected/resource/23",
				TargetMethod: http.MethodGet,
			},
			ExpectedHTTPCode: http.StatusUnauthorized,
		},
	}
}

func TestMultipleMethodsAreSupported(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost}
	cfg := middleware.NewAuthConfig().
		WithMethods(methods).
		WithAuth(userAuth)
	cfg.WithPathRegex("^/protected.*")

	mid := middleware.AuthChecker(cfg)

	for _, m := range methods {
		req := httptest.NewRequest(m, "/protected", nil)
		req.SetBasicAuth("user", "password")
		rec := httptest.NewRecorder()
		mid(nullHandler).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	}
}

func TestAuthCheckerPassOnNil(t *testing.T) {
	mid := middleware.AuthChecker(nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mid(nullHandler).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
}
