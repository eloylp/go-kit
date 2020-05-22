// +build unit

package fanout_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eloylp/kit/flow/fanout"
	"github.com/eloylp/kit/flow/fanout/fanouttest"
)

func TestBufferedFanOut_Subscribe_ElementsAreSentToSubscribers(t *testing.T) {
	elems := 3
	maxBuffLen := 10
	fo := fanouttest.BufferedFanOut(maxBuffLen, time.Now)
	ch, _, _ := fo.Subscribe()
	fanouttest.Populate(fo, elems)
	var chOk bool
	var slot *fanout.Slot
	for i := 0; i < elems; i++ {
		slot, chOk = <-ch
		item, ok := slot.Elem.([]byte)
		assert.True(t, ok, "wanted []byte type got %T", item)
		want := "d" + fmt.Sprint(i)
		got := string(item)
		assert.Equal(t, want, got, "error listening subscribed elements, wanted was %q but got %q", want, got)
	}
	assert.True(t, chOk, "channel remains open for future consumes")
}

func TestBufferedFanOut_Subscribe_ReturnValues(t *testing.T) {
	elems := 3
	maxBuffLen := 10
	fo := fanouttest.BufferedFanOut(maxBuffLen, time.Now)
	ch, uuid, _ := fo.Subscribe()
	fanouttest.Populate(fo, elems)
	assert.NotEmpty(t, uuid, "want a uuid not an empty string")
	assert.NotNil(t, ch, "want a channel")
	_, ok := <-ch
	assert.True(t, ok, "want a open channel")
}

func TestBufferedFanOut_Unsubscribe(t *testing.T) {
	maxBuffLen := 10
	fo := fanouttest.BufferedFanOut(maxBuffLen, time.Now)
	// Adds one extra subscriber for test hardening.
	_, _, _ = fo.Subscribe() //nolint:dogsled
	ch, uuid, _ := fo.Subscribe()
	elems := 3
	fanouttest.Populate(fo, elems)
	err := fo.Unsubscribe(uuid)
	assert.NoError(t, err)
	assert.Equal(t, 1, fo.SubscribersLen())
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

func TestBufferedFanOut_Unsubscribe_WithCancelFunc(t *testing.T) {
	maxBuffLen := 10
	fo := fanouttest.BufferedFanOut(maxBuffLen, time.Now)
	// Adds one extra subscriber for test hardening.
	_, _, _ = fo.Subscribe() //nolint:dogsled
	ch, _, cancel := fo.Subscribe()
	elems := 3
	fanouttest.Populate(fo, elems)
	err := cancel()
	assert.NoError(t, err)
	assert.Equal(t, 1, fo.SubscribersLen())
	// exhaust channel
	var count int
	for range ch {
		if count == 2 {
			break
		}
		count++
	}
	_, ok := <-ch
	assert.False(t, ok, "want channel closed after Unsubscribe and consumed")
}

func TestBufferedFanOut_Unsubscribe_NotFound(t *testing.T) {
	fo := fanouttest.BufferedFanOut(10, time.Now)
	_, _, cancel := fo.Subscribe()
	fo.Reset()
	err := cancel()
	assert.IsType(t, fanout.ErrSubscriberNotFound, err, "wanted fanout.ErrSubscriberNotFound got %T", err)
}

func TestBufferedFanOut_Reset(t *testing.T) {
	maxBuffLen := 10
	fo := fanouttest.BufferedFanOut(maxBuffLen, time.Now)
	ch, _, _ := fo.Subscribe()
	fo.Add([]byte("dd"))
	fo.Reset()
	assert.Equal(t, 0, fo.SubscribersLen(), "no subscribers expected after reset")
	// Check channel is closed after consumption
	<-ch
	_, ok := <-ch
	assert.False(t, ok)
}

func TestBufferedFanOut_Add_NoActiveSubscriberDoesntBlock(t *testing.T) {
	maxBuffLen := 10
	fo := fanouttest.BufferedFanOut(maxBuffLen, time.Now)
	ch, _, _ := fo.Subscribe() // this subscriber will try to block the entire system
	elems := 3                 // "d + index" elems (continue reading comments ...)
	fanouttest.Populate(fo, elems)
	limitValueForBlocking := maxBuffLen // So we will overwrite entire channel with new data "dn" (limit value of maxBuffLen)
	testEnd := make(chan struct{}, 1)
	go func() {
		for i := 0; i < limitValueForBlocking; i++ {
			// "dn + index" will mark new data segments that may override old ones in factory "d + index".
			fo.Add([]byte("dn" + strconv.Itoa(i)))
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

func TestBufferedFanOut_Status_Count(t *testing.T) {
	maxBuffLen := 10
	fo := fanouttest.BufferedFanOut(maxBuffLen, time.Now)
	_, uid1, _ := fo.Subscribe()
	ch2, uid2, _ := fo.Subscribe()
	fo.Add([]int{1, 1})
	fo.Add([]int{2, 2})
	<-ch2
	want := fanout.Status{
		uid1: 2,
		uid2: 1,
	}
	assert.Equal(t, want, fo.Status())
}

func TestBufferedFanOut_Status_Unsubscribe(t *testing.T) {
	maxBuffLen := 10
	fo := fanouttest.BufferedFanOut(maxBuffLen, time.Now)
	_, _, cancel := fo.Subscribe()
	_, uid2, _ := fo.Subscribe()
	fo.Add([]int{1, 1})
	fo.Add([]int{2, 2})
	err := cancel()
	assert.NoError(t, err)
	want := fanout.Status{
		uid2: 2,
	}
	assert.Equal(t, want, fo.Status())
}
