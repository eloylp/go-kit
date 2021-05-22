// Package middleware will hold common use HTTP middlewares.
// The middlewares are totally compliant with the standard lib interfaces
// and with most of the routers out there.
package middleware

import (
	"net/http"
)

// Middleware accepts the next http.Handler as parameter
// and returns current one that may modify request/writer
// and finally calls the handler passed as parameter.
type Middleware func(h http.Handler) http.Handler

// InFrontOf will take the handler as first parameter.
// The variadic part of function accepts any number of middlewares
// that will be called in the passed order.
// Beware that the handler will always be called as the
// last element of the chain, after all the middlewares are called.
func InFrontOf(h http.Handler, m ...Middleware) http.Handler {
	for j := len(m) - 1; j >= 0; j-- {
		h = m[j](h)
	}
	return h
}
