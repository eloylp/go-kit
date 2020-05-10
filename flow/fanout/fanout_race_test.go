// +build race

package fanout_test

import (
	"testing"
	"time"

	"github.com/eloylp/go-kit/flow/fanout/fanouttest"
)

func TestBufferedFanOut_AddElem_SupportsRace(t *testing.T) {
	fo := fanouttest.PopulatedBufferedFanOut(t, 3, 3, 5)
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
