// +build concurrent

package fanout_test

import (
	"fmt"
	"github.com/eloylp/go-kit/flow/fanout"
	"github.com/eloylp/go-kit/flow/fanout/fanouttest"
	"testing"
	"time"
)

func TestBufferedFanOut_SupportsRace(t *testing.T) {
	fo := fanouttest.BufferedFanOut(5, time.Now)
	timer := time.NewTimer(time.Second * 60)
	cancels := make(chan fanout.CancelFunc, 10)
	// Add status vector
	go func() {
		for {
			time.Sleep(time.Second * 2)
			t.Log(fo.Status())
			fmt.Println("sd")
		}
	}()
	// Subscribe vector
	go func() {
		for {
			subs, _, cancel := fo.Subscribe()
			cancels <- cancel
			go func() {
				for {
					<-subs
				}
			}()
		}
	}()
	// Unsubscribe vector
	go func() {
		for c := range cancels {
			time.Sleep(time.Millisecond * 700)
			c()
		}
	}()
	// Add elem vector
	go func() {
		for {
			fo.AddElem([]byte("data"))
		}
	}()
	// Add reset vector
	go func() {
		for {
			time.Sleep(time.Second * 5)
			fo.Reset()
		}
	}()
	// Add subs len vector
	go func() {
		for {
			time.Sleep(time.Millisecond * 500)
			fo.SubscribersLen()
		}
	}()
	<-timer.C
}
