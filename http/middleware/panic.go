package middleware

import "net/http"

// PanicHandlerFunc represents the function
// that will handle the panic. This is an
// user provided func.
type PanicHandlerFunc func(v interface{})

// PanicHandler its a middleware that will
// intercept panics and act according to an
// user provided function (See PanicHandlerFunc).
//
// It should be invoked before any other handler
// logic and other middlewares, protecting them
// from panics.
//
// Its only able to intercept panics that happens
// in the same goroutine.
func PanicHandler(panicHandler PanicHandlerFunc) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					panicHandler(err)
				}
			}()
			handler.ServeHTTP(w, r)
		})
	}
}
