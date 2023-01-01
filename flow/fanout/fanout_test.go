package fanout_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/kit/flow/fanout"
)

func TestBufferedFanOut_Subscribe_ElementsAreSentToSubscribers(t *testing.T) {
	elems := 3
	maxBuffLen := 10
	fo := fanout.NewBufferedFanOut[[]byte](maxBuffLen, time.Now)

	ch, _ := fo.Subscribe()

	fo.Add([]byte(strconv.Itoa(1)))
	fo.Add([]byte(strconv.Itoa(2)))
	fo.Add([]byte(strconv.Itoa(3)))

	var chOk bool
	var slot *fanout.Slot[[]byte]

	for i := 1; i <= elems; i++ {
		slot, chOk = <-ch
		want := fmt.Sprint(i)
		got := string(slot.Elem)
		assert.Equal(t, want, got, "error listening subscribed elements, wanted was %q but got %q", want, got)
	}
	assert.True(t, chOk, "channel remains open for future consumes")
}

func TestBufferedFanOut_Subscribe_ReturnValues(t *testing.T) {
	maxBuffLen := 10

	fo := fanout.NewBufferedFanOut[[]byte](maxBuffLen, time.Now)
	ch, _ := fo.Subscribe()

	fo.Add([]byte(strconv.Itoa(1)))
	fo.Add([]byte(strconv.Itoa(2)))
	fo.Add([]byte(strconv.Itoa(3)))

	assert.NotNil(t, ch, "want a channel")
	_, ok := <-ch
	assert.True(t, ok, "want a open channel")
}

func TestBufferedFanOut_Unsubscribe(t *testing.T) {
	maxBuffLen := 10
	fo := fanout.NewBufferedFanOut[[]byte](maxBuffLen, time.Now)
	// Adds one extra subscriber for test hardening.
	_, _ = fo.Subscribe() //nolint:dogsled
	ch, unsubscribe := fo.Subscribe()

	fo.Add([]byte(strconv.Itoa(1)))
	fo.Add([]byte(strconv.Itoa(2)))

	err := unsubscribe()
	require.NoError(t, err)
	assert.Equal(t, 1, fo.SubscribersLen())
	// exhaust channel
	var count int
	for range ch {
		count++
		if count == 2 {
			break
		}
	}
	_, ok := <-ch
	assert.False(t, ok, "want channel closed after unsubscribe and consumed")
}

func TestBufferedFanOut_Unsubscribe_NotFound(t *testing.T) {
	maxBuffLen := 10
	fo := fanout.NewBufferedFanOut[[]byte](maxBuffLen, time.Now)
	_, cancel := fo.Subscribe()
	fo.Reset()
	err := cancel()
	assert.IsType(t, fanout.ErrSubscriberNotFound, err, "wanted fanout.ErrSubscriberNotFound got %T", err)
}

func TestBufferedFanOut_Reset(t *testing.T) {
	maxBuffLen := 10
	fo := fanout.NewBufferedFanOut[[]byte](maxBuffLen, time.Now)

	ch, _ := fo.Subscribe()
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
	fo := fanout.NewBufferedFanOut[[]byte](maxBuffLen, time.Now)
	ch, _ := fo.Subscribe() // this subscriber will try to block the entire system

	fo.Add([]byte(strconv.Itoa(1)))
	fo.Add([]byte(strconv.Itoa(2)))
	fo.Add([]byte(strconv.Itoa(3)))

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
		assert.Contains(t, string(s.Elem), "dn",
			"The are elements in subs channel that needs to be discarded")
	}
}

func TestBufferedFanOut_Status_Count_Aggregated(t *testing.T) {
	maxBuffLen := 10
	fo := fanout.NewBufferedFanOut[[]int](maxBuffLen, time.Now)

	_, _ = fo.Subscribe()
	_, _ = fo.Subscribe()

	fo.Add([]int{1, 1})
	fo.Add([]int{2, 2})

	want := fanout.Status{
		"": 4,
	}
	assert.Equal(t, want, fo.Status())
}

func TestBufferedFanOut_Status_Count(t *testing.T) {
	maxBuffLen := 10
	fo := fanout.NewBufferedFanOut[[]int](maxBuffLen, time.Now)

	_, _ = fo.SubscribeWith("a")
	ch2, _ := fo.SubscribeWith("b")
	fo.Add([]int{1, 1})
	fo.Add([]int{2, 2})
	<-ch2
	want := fanout.Status{
		"a": 2,
		"b": 1,
	}
	assert.Equal(t, want, fo.Status())
}

func TestBufferedFanOut_Status_Unsubscribe(t *testing.T) {
	maxBuffLen := 10
	fo := fanout.NewBufferedFanOut[[]int](maxBuffLen, time.Now)

	_, cancel := fo.SubscribeWith("a")
	_, _ = fo.SubscribeWith("b")
	fo.Add([]int{1, 1})
	fo.Add([]int{2, 2})
	err := cancel()
	require.NoError(t, err)
	want := fanout.Status{
		"b": 2,
	}
	assert.Equal(t, want, fo.Status())
}
