// Package middleware will hold common use HTTP middlewares.
// The middlewares are totally compliant with the standard lib interfaces
// and with most of the routers out there.
package middleware

import (
	"net/http"
)

// Middleware is the type that the middleware closures
// must return.
type Middleware func(handler http.Handler) http.Handler
