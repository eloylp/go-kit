package flow_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.eloylp.dev/kit/flow"
)

// This is a racy test. The target is to find nasty data races in the code.
// It NEEDS to be executed with the race detector enabled.
// See https://go.dev/blog/race-detector .
//
// It consists on trying to execute all available code paths
// at the same time, during certain time, exercising all the
// locks and internal data structures.
//
// Its important to remember to add all the new added features/code paths here.

func TestFanout_SupportsRace(t *testing.T) {

	ctx, testCancel := context.WithCancel(context.Background())

	fo := flow.NewFanout[int](20)

	var wg sync.WaitGroup

	// Add status code path
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(600 * time.Millisecond)
				fo.Status()
			}
		}
	}()

	cancels := make(chan flow.CancelFunc, 100_000)

	// Subscriber code path (subscribes and consumes)
	subscribersVector(ctx, &wg, fo, cancels)

	// Unsubscribe code path
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-cancels:
				time.Sleep(500 * time.Millisecond)
				c()
			}
		}
	}()

	// Add elem code path
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fo.Add(1)
				time.Sleep(100 * time.Microsecond)
			}
		}
	}()

	// Add reset code path
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 5)
		fo.Reset()
		subscribersVector(ctx, &wg, fo, cancels)
	}()

	// Add subs active code path
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fo.ActiveSubscribers()
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	// Add subs len code path
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fo.SubscribersLen()
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	// Enough ! lets stop all things
	time.AfterFunc(10*time.Second, func() {
		fo.Reset()
		testCancel()
	})
	// And properly wait for them.
	wg.Wait()
}

func subscribersVector[T any](ctx context.Context, wg *sync.WaitGroup, fo *flow.Fanout[T], cancels chan flow.CancelFunc) {
	for i := 0; i < 10; i++ {
		consume, cancel := fo.Subscribe()
		cancels <- cancel
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					cancel()
					return
				default:
					go consume()
				}
			}
		}()
	}
}
