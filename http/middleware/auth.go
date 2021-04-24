package middleware

import (
	"net/http"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// Authorization represent the user/encrypted-password table.
// The keys are reserved for the user and the values for their respective
// bcrypt encrypted passwords.
type Authorization map[string]string

// AuthChecker is a middleware to prevent/allow access to
// a specific combination of HTTP method/path. Read authMiddlewareConfig
// for more information about how to configure this middleware.
// This middleware should be the executed just before your business logic.
func AuthChecker(cfg *authMiddlewareConfig) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == cfg.method {
				for _, p := range cfg.patterns {
					if !p.MatchString(r.URL.String()) {
						continue
					}
					user, pass, ok := r.BasicAuth()
					if !ok {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte("Unauthorized"))
						return
					}
					passHash, ok := cfg.authorizations[user]
					if !ok {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte("Unauthorized"))
						return
					}
					err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(pass))
					if err != nil {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte("Unauthorized"))
						return
					}
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

type authMiddlewareConfig struct {
	method         string
	patterns       []*regexp.Regexp
	authorizations Authorization
}

// NewAuthMiddlewareConfig returns a builder pattern that
// will help to build a config for setup the AuthChecker middleware.
// you must follow the fluent interface to set all the fields in the
// struct.
func NewAuthMiddlewareConfig() *authMiddlewareConfig { //nolint:golint
	return &authMiddlewareConfig{}
}

// WithMethod sets the HTTP method where all the protected endpoints will reside.
func (a *authMiddlewareConfig) WithMethod(m string) *authMiddlewareConfig {
	a.method = m
	return a
}

// WithPathRegex appends a regex pattern to be matched with the input path of the user.
// The HTTP paths that matches this regexes will be protected. This method can be called
// multiple times to add multiple regex.
func (a *authMiddlewareConfig) WithPathRegex(p string) *authMiddlewareConfig {
	a.patterns = append(a.patterns, regexp.MustCompile(p))
	return a
}

// WithAuth maps the user/password permissions. See Authorization struct type for more
// information.
func (a *authMiddlewareConfig) WithAuth(auth Authorization) *authMiddlewareConfig {
	a.authorizations = auth
	return a
}
