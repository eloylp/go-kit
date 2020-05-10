// +build unit

package fanout_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eloylp/go-kit/flow/fanout"
	"github.com/eloylp/go-kit/flow/fanout/fanouttest"
)

func TestBufferedFanOut_Subscribe_ElementsAreSentToSubscribers(t *testing.T) {
	elems := 3
	maxElems := 3
	maxSubsBuffSize := 10
	fo := fanouttest.PopulatedBufferedFanOut(t, elems, maxElems, maxSubsBuffSize)
	ch, _ := fo.Subscribe()
	err := fo.AddElem([]byte("d3")) // Extra (4th) data that is removed by excess
	assert.NoError(t, err)
	var chOk bool
	var slot *fanout.Slot
	for i := 0; i < elems+1; i++ {
		slot, chOk = <-ch
		item, ok := slot.Elem.([]byte)
		assert.True(t, ok, "wanted []byte type got %T", item)
		want := "d" + fmt.Sprint(i)
		got := string(item)
		// We check that all data is present, even removed by excess
		assert.Equal(t, want, got, "error listening subscribed elements, wanted was %q but got %q", want, got)
	}
	assert.True(t, chOk, "channel remains open for future consumes")
}

func TestBufferedFanOut_Subscribe_ReturnValues(t *testing.T) {
	elems := 3
	maxElems := 3
	maxSubsBuffSize := 10
	fo := fanouttest.PopulatedBufferedFanOut(t, elems, maxElems, maxSubsBuffSize)
	ch, uuid := fo.Subscribe()
	assert.NotEmpty(t, uuid, "want a uuid not an empty string")
	assert.NotNil(t, ch, "want a channel")
	_, ok := <-ch
	assert.True(t, ok, "want a open channel")
}

func TestBufferedFanOut_Unsubscribe(t *testing.T) {
	elems := 3
	maxElems := 3
	maxSubsBuffSize := 10
	fo := fanouttest.PopulatedBufferedFanOut(t, elems, maxElems, maxSubsBuffSize)
	// Adds one extra subscriber for test hardening.
	_, _ = fo.Subscribe()
	ch, uuid := fo.Subscribe()
	err := fo.Unsubscribe(uuid)
	assert.NoError(t, err)
	assert.Equal(t, 1, fo.Subscribers())

	// exhaust channel
	var count int
	for range ch {
		if count == 2 {
			break
		}
		count++
	}
	_, ok := <-ch
	assert.False(t, ok, "want channel closed after unsubscribe and consumed")
}

func TestBufferedFanOut_Unsubscribe_NotFound(t *testing.T) {
	elems := 3
	maxElems := 3
	maxSubsBuffSize := 10
	fo := fanouttest.PopulatedBufferedFanOut(t, elems, maxElems, maxSubsBuffSize)
	err := fo.Unsubscribe("A1234")
	assert.IsType(t, fanout.ErrSubscriberNotFound, err, "wanted fanout.ErrSubscriberNotFound got %T", err)
}

func TestBufferedFanOut_Reset(t *testing.T) {
	elems := 0
	maxElems := 3
	maxSubsBuffSize := 10
	fo := fanouttest.PopulatedBufferedFanOut(t, elems, maxElems, maxSubsBuffSize)
	ch, _ := fo.Subscribe()
	err := fo.AddElem([]byte("dd"))
	assert.NoError(t, err)
	fo.Reset()
	assert.Equal(t, 0, fo.Subscribers(), "no subscribers expected after reset")
	assert.Equal(t, 0, fo.Length(), "no elems expected after reset")
	// Check channel is closed after consumption
	<-ch
	_, ok := <-ch
	assert.False(t, ok)
}

func TestNewBufferedFanOut_AddItem_OldItemsClear(t *testing.T) {
	elems := 3
	maxElems := 3
	maxSubsBuffSize := 10
	fo := fanouttest.PopulatedBufferedFanOut(t, elems, maxElems, maxSubsBuffSize)
	err := fo.AddElem([]byte("d4"))
	assert.NoError(t, err)
	want := 3
	got := fo.Length()
	assert.Equal(t, want, got, "want %v resultant elems got %v", want, got)
}

func TestBufferedFanOut_AddItem_NoActiveSubscriberDoesntBlock(t *testing.T) {
	elems := 3 // "d + index" elems (continue reading comments ...)
	maxElems := 3
	maxSubsBuffSize := 10
	fo := fanouttest.PopulatedBufferedFanOut(t, elems, maxElems, maxSubsBuffSize)
	ch, _ := fo.Subscribe()     // this subscriber will try to block the entire system
	limitValueForBlocking := 11 // So we will exceed by one (limit value of maxSubsBuffSize)

	testEnd := make(chan struct{}, 1)
	go func() {
		for i := 0; i < limitValueForBlocking; i++ {
			// "dn + index" will mark new data segments that may override old ones in factory "d + index".
			err := fo.AddElem([]byte("dn" + strconv.Itoa(i)))
			assert.NoError(t, err)
		}
		testEnd <- struct{}{}
	}()
	select {
	case <-time.NewTimer(2 * time.Second).C: // Will break test if its blocking.
		t.Error("exceeded wait time. May subscribers are blocking the buffer")
	case <-testEnd:
		t.Log("successfully inserted elems without active subscribers")
	}
	fo.Reset() // this will close the subscriber channel, allowing us to follow with the next check.

	// Now check that all discarded elements in subscriber buffer are old, from the factory.
	// This will be done by checking all data in subscriber channels contains "dn + index".
	for s := range ch {
		content, ok := s.Elem.([]byte)
		assert.True(t, ok, "wanted type []byte got %T", content)
		assert.Contains(t, string(content), "dn",
			"The are elements in subs channel that needs to be discarded")
	}
}
