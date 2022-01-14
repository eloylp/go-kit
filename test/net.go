package test

import (
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"
)

// WaitTCPService will try to connect to the provided TCP socket.
// It will keep trying in the specified intervals for the specified
// amount of time.
//
// It will declare the test as Failed, stopping the execution on fail.
func WaitTCPService(t *testing.T, addr string, interval, maxWait time.Duration) {
	t.Helper()
	ctx, cancl := context.WithTimeout(context.Background(), maxWait)
	defer cancl()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("WaitTCPService(): %v", ctx.Err())
		default:
			con, conErr := net.Dial("tcp", addr)
			if conErr == nil {
				_ = con.Close()
				return
			}
			time.Sleep(interval)
		}
	}
}

// WaitTLSService will try to connect to the provided TLS socket.
// It will keep trying in the specified intervals for the specified
// amount of time.
//
// It will not do any X509 certificate verification as this is intended
// to just try connectivity without getting SSL certificate checks errors
// in logs.
//
// It will declare the test as Failed, stopping the execution on fail.
func WaitTLSService(t *testing.T, addr string, interval, maxWait time.Duration) {
	t.Helper()
	ctx, cancl := context.WithTimeout(context.Background(), maxWait)
	defer cancl()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("WaitTLSService(): %v", ctx.Err())
		default:
			con, conErr := tls.Dial("tcp", addr, &tls.Config{
				InsecureSkipVerify: true,
			})
			if conErr == nil {
				_ = con.Close()
				return
			}
			time.Sleep(interval)
		}
	}
}
