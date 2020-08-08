package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// WithOSSignals will set up a OS signal watcher that will gracefully
// shutdown the server when a proper signal is received.
// If an error occurs, and a func(err error) is passed as last argument,
// that will be used for handling the error. An already incremented wait group
// is needed for wait for the shutdown method to finish.
func WithOSSignals(s *http.Server, timeout time.Duration, wg *sync.WaitGroup, errHandler func(err error)) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer wg.Done()
		<-signals
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil && errHandler != nil {
			errHandler(err)
		}
	}()
}
