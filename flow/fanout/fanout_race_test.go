// +build race

package fanout_test

import (
	"github.com/eloylp/go-kit/flow/fanout/fanouttest"
	"testing"
	"time"
)

func TestBufferedFanOut_AddElem_SupportsRace(t *testing.T) {
	fo := fanouttest.BufferedFanOut(5)
	subs, _ := fo.Subscribe()
	go func() {
		for {
			<-subs
		}
	}()
	timer := time.NewTimer(time.Second * 10)
loop:
	for {
		select {
		case <-timer.C:
			break loop
		default:
			go fo.AddElem([]byte("d")) //nolint:errcheck
		}
	}
}
