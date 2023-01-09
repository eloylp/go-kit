package middleware

import (
	"net/http"
)

type DefaultHeaders map[string]string

func (d DefaultHeaders) Set(k, v string) {
	d[k] = v
}

// ResponseHeaders middleware will setup the
// default response headers.
//
// It will override any other previous settled
// headers in the http.ResponseWriter. Downstream
// handlers can still modify the default headers.
//
// Users must take into account where this middleware
// its placed in the handler chain. i.e adding this
// middleware too early, can cause other middlewares to
// overwrite the wanted defaults. And vice-versa, putting
// this too late, will override other middleware headers.
func ResponseHeaders(headers DefaultHeaders) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			h.ServeHTTP(w, r)
		})
	}
}
