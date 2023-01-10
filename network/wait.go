package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// WaitTCPService will try to connect to the provided TCP socket.
// It will keep trying in the specified intervals for the specified
// amount of time.
func WaitTCPService(ctx context.Context, addr string, interval time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("WaitTCPService(): %v", ctx.Err())
		default:
			conn, err := net.Dial("tcp", addr)
			if err == nil {
				_ = conn.Close()
				return nil
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
func WaitTLSService(ctx context.Context, addr string, interval time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("WaitTLSService(): %v", ctx.Err())	
		default:
			conn, err := tls.Dial("tcp", addr, &tls.Config{
				InsecureSkipVerify: true,
			})
			if err == nil {
				_ = conn.Close()
				return nil
			}
			time.Sleep(interval)
		}
	}
}
