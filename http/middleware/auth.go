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

// AuthConfigFunc is the only one input parameter of the
// AuthChecker middleware. It must return an []*AuthConfig,
// That can come from any source, like a database.
//
// If the source is going to be a database, its encouraged,
// to some kind of caching, as this is going to be executed
// for EVERY request.
type AuthConfigFunc func() []*AuthConfig

// AuthChecker is a middleware to prevent/allow access to
// a specific combination of HTTP method/path. Read AuthConfig
// for more information about how to configure this middleware.
// This middleware should be the executed just before your business logic.
func AuthChecker(cfgFunc AuthConfigFunc) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfgFunc == nil {
				h.ServeHTTP(w, r)
				return
			}
			for _, cfg := range cfgFunc() {
				if isConfiguredMethod(r.Method, cfg.Method) {
					for _, p := range cfg.Patterns {
						if !p.MatchString(r.URL.String()) {
							continue
						}
						user, pass, ok := r.BasicAuth()
						if !ok {
							w.WriteHeader(http.StatusUnauthorized)
							_, _ = w.Write([]byte("Unauthorized"))
							return
						}
						passHash, ok := cfg.Authorizations[user]
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
			}
			h.ServeHTTP(w, r)
		})
	}
}

func isConfiguredMethod(requestMethod string, cfgMethods []string) bool {
	for _, m := range cfgMethods {
		if m == requestMethod {
			return true
		}
	}
	return false
}

type AuthConfig struct {
	Method         []string
	Patterns       []*regexp.Regexp
	Authorizations Authorization
}

// NewAuthConfig returns a builder pattern that
// will help to build a config for setup the AuthChecker middleware.
// you must follow the fluent interface to set all the fields in the
// struct.
func NewAuthConfig() *AuthConfig { //nolint:golint
	return &AuthConfig{}
}

// WithMethods sets the HTTP methods where all the protected endpoints will reside.
func (a *AuthConfig) WithMethods(m []string) *AuthConfig {
	a.Method = m
	return a
}

// WithPathRegex appends a regex pattern to be matched with the input path of the user.
// The HTTP paths that matches this regexes will be protected. This method can be called
// multiple times to add multiple regex.
func (a *AuthConfig) WithPathRegex(p string) *AuthConfig {
	a.Patterns = append(a.Patterns, regexp.MustCompile(p))
	return a
}

// WithAuth maps the user/password permissions. See Authorization struct type for more
// information.
func (a *AuthConfig) WithAuth(auth Authorization) *AuthConfig {
	a.Authorizations = auth
	return a
}
