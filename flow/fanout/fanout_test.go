package fanout_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/kit/flow/fanout"
)

func TestBufferedFanOut_Subscribe_ElementsAreSentToSubscribers(t *testing.T) {
	fo := fanout.NewBufferedFanOut[string](10, time.Now)

	ch, _ := fo.Subscribe()

	fo.Add("1")
	fo.Add("2")
	fo.Add("3")

	var chOk bool
	var slot *fanout.Slot[string]

	for i := 1; i <= 3; i++ {
		slot, chOk = <-ch
		want := fmt.Sprint(i)
		got := string(slot.Elem)
		assert.Equal(t, want, got, "error listening subscribed elements, wanted was %q but got %q", want, got)
	}
	assert.True(t, chOk, "channel remains open for future consumes")
}

func TestBufferedFanOut_Subscribe_ReturnValues(t *testing.T) {
	fo := fanout.NewBufferedFanOut[string](10, time.Now)
	ch, _ := fo.Subscribe()

	fo.Add("a")
	fo.Add("b")
	fo.Add("c")

	assert.NotNil(t, ch, "want a channel")
	_, ok := <-ch
	assert.True(t, ok, "want a open channel")
}

func TestBufferedFanOut_Unsubscribe(t *testing.T) {
	fo := fanout.NewBufferedFanOut[string](10, time.Now)
	// Adds one extra subscriber for test hardening.
	_, _ = fo.Subscribe() //nolint:dogsled
	ch, unsubscribe := fo.Subscribe()

	fo.Add("a")
	fo.Add("b")

	err := unsubscribe()
	require.NoError(t, err)
	assert.Equal(t, 1, fo.ActiveSubscribers())
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
	fo := fanout.NewBufferedFanOut[string](10, time.Now)
	_, cancel := fo.Subscribe()
	fo.Reset()
	err := cancel()
	assert.IsType(t, fanout.ErrSubscriberNotFound, err, "wanted fanout.ErrSubscriberNotFound got %T", err)
}

func TestBufferedFanOut_Reset(t *testing.T) {
	fo := fanout.NewBufferedFanOut[string](10, time.Now)

	ch, _ := fo.Subscribe()
	fo.Add("dd")
	fo.Reset()
	assert.Equal(t, 0, fo.ActiveSubscribers(), "no subscribers expected after reset")
	// Check channel is closed after consumption
	<-ch
	_, ok := <-ch
	assert.False(t, ok)
}

func TestBufferedFanOut_Add_NoActiveSubscriberDoesntBlock(t *testing.T) {
	maxBuffLen := 10
	fo := fanout.NewBufferedFanOut[string](maxBuffLen, time.Now)
	ch, _ := fo.Subscribe() // this subscriber will try to block the entire system

	fo.Add("d1")
	fo.Add("d2")
	fo.Add("d3")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for i := 0; i < maxBuffLen; i++ { // We will overwrite entire channel with new data "nd" (limit value of maxBuffLen)
			// "nd + index" will mark new data segments that may override old ones in factory "d + index".
			fo.Add("dn" + strconv.Itoa(i))
		}
		cancel()
	}()
	select {
	case <-time.NewTimer(2 * time.Second).C: // Will break test if its blocking.
		t.Error("exceeded wait time. May subscribers are blocking the buffer")
	case <-ctx.Done():
		t.Log("successfully inserted elems without active subscribers")
	}
	fo.Reset() // this will close the subscriber channel, allowing us to follow with the next check.

	// Now check that all discarded elements in subscriber buffer are old, from the factory.
	// This will be done by checking all data in subscriber channels contains "nd + index".
	for s := range ch {
		assert.Contains(t, string(s.Elem), "dn",
			"The are elements in subs channel that needs to be discarded")
	}
}

func TestBufferedFanOut_Status_Count_Aggregated(t *testing.T) {
	fo := fanout.NewBufferedFanOut[int](10, time.Now)

	_, _ = fo.Subscribe()
	_, _ = fo.Subscribe()

	fo.Add(1)
	fo.Add(2)

	want := fanout.Status{
		"": 4,
	}
	assert.Equal(t, want, fo.Status())
}

func TestBufferedFanOut_Status_Count(t *testing.T) {
	fo := fanout.NewBufferedFanOut[int](10, time.Now)

	_, _ = fo.SubscribeWith("a")
	ch2, _ := fo.SubscribeWith("b")
	fo.Add(1)
	fo.Add(2)
	<-ch2
	want := fanout.Status{
		"a": 2,
		"b": 1,
	}
	assert.Equal(t, want, fo.Status())
}

func TestBufferedFanOut_Status_Unsubscribe(t *testing.T) {
	fo := fanout.NewBufferedFanOut[int](10, time.Now)

	_, cancel := fo.SubscribeWith("a")
	_, _ = fo.SubscribeWith("b")
	fo.Add(1)
	fo.Add(2)
	err := cancel()
	require.NoError(t, err)
	want := fanout.Status{
		"b": 2,
	}
	assert.Equal(t, want, fo.Status())
}

func TestSubscribersStoreReuse(t *testing.T) {
	fo := fanout.NewBufferedFanOut[int](10, time.Now)

	fo.Subscribe() // 1
	fo.Subscribe() // 2

	// Cancelling the 3th subscription.
	_, cancel := fo.Subscribe() // 3
	cancel()

	fo.Subscribe() // 4 This subscriber should be allocated in same place of previous canceled subscription (3).

	assert.Equal(t, 3, fo.SubscribersLen(), "It should not reserve more subscriber slots, if we already have empty ones.")

}

func TestSubscribersStoreGrows(t *testing.T) {
	fo := fanout.NewBufferedFanOut[int](10, time.Now)

	fo.Subscribe()
	fo.Subscribe()
	fo.Subscribe()

	assert.Equal(t, 3, fo.SubscribersLen(), "Subscriber len should grow linearly")

}
