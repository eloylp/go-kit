package middleware

import (
	"net/http"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// AuthConfig are the elements the AuthConfigFunc
// needs to return in order to configure the
// AuthChecker middleware.
//
// Having the possibility of returning multiple of this
// configurations in AuthConfigFunc allows configuring
// different user accesses for maybe different set of
// endpoints.
type AuthConfig struct {
	Methods []string
	Paths   []*regexp.Regexp
	Auth    Authorization
}

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
				if isConfiguredMethod(r.Method, cfg.Methods) {
					for _, p := range cfg.Paths {
						if !p.MatchString(r.URL.String()) {
							continue
						}
						user, pass, ok := r.BasicAuth()
						if !ok {
							w.WriteHeader(http.StatusUnauthorized)
							_, _ = w.Write([]byte("Unauthorized"))
							return
						}
						passHash, ok := cfg.Auth[user]
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

// NewAuthConfig returns a builder pattern that
// helps to build a config for setup the
// AuthChecker middleware.
func NewAuthConfig() *AuthConfig { //nolint:golint
	return &AuthConfig{}
}

// WithMethods configures the methods that are going
// to be protected.
func (a *AuthConfig) WithMethods(m []string) *AuthConfig {
	a.Methods = m
	return a
}

// WithPathRegex appends a regex pattern to be matched
// with the input URL of the user.
//
// The HTTP paths that matches this regex's will be protected.
// This method can be called multiple times to add multiple regex.
func (a *AuthConfig) WithPathRegex(p string) *AuthConfig {
	a.Paths = append(a.Paths, regexp.MustCompile(p))
	return a
}

// WithAuth maps the user/password permissions. See Authorization
// for more information.
func (a *AuthConfig) WithAuth(auth Authorization) *AuthConfig {
	a.Auth = auth
	return a
}

// AllMethods is a helper that provides all available
// HTTP methods.
//
// Useful whenever we want to configure ALL methods.
func AllMethods() []string {
	return []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodHead,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}
}
