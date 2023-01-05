package fanout_test

import (
	"sync"
	"testing"
	"time"

	"go.eloylp.dev/kit/flow/fanout"
)

func TestBufferedFanOut_SupportsRace(t *testing.T) {
	fo := fanout.NewBufferedFanOut[[]byte](5, time.Now)
	cancels := make(chan fanout.CancelFunc, 10)
	var wg sync.WaitGroup
	t.Log("starting racy test ...")
	// Add status vector
	go func() {
		for {
			fo.Status()
		}
	}()
	// Subscribe vector
	wg.Add(1)
	go func() {
		for i := 0; i < 8000; i++ {
			consume, cancel := fo.Subscribe()
			cancels <- cancel
			go func() {
				for {
					time.Sleep(time.Millisecond * 300)
					consume()
				}
			}()
		}
		close(cancels)
		wg.Done()
	}()
	// Unsubscribe vector
	wg.Add(1)
	go func() {
		for c := range cancels {
			_ = c()
		}
		wg.Done()
	}()
	// Add elem vector
	go func() {
		for {
			fo.Add([]byte("data"))
		}
	}()
	// Add reset vector
	wg.Add(1)
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(time.Millisecond * 500)
			fo.Reset()
		}
		wg.Done()
	}()
	// Add subs active vector
	go func() {
		for {
			fo.ActiveSubscribers()
		}
	}()
	// Add subs len vector
	go func() {
		for {
			fo.SubscribersLen()
		}
	}()
	wg.Wait()
}
