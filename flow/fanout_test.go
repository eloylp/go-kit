package flow_test

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/kit/flow"
)

func TestFanout_Subscribe_ElementsAreSentToSubscribers(t *testing.T) {
	fo := flow.NewBufferedFanOut[string](10)

	consume, _ := fo.Subscribe()

	fo.Add("1")
	fo.Add("2")
	fo.Add("3")

	var err error
	var slot *flow.Slot[string]

	for i := 1; i <= 3; i++ {
		slot, err = consume()
		want := fmt.Sprint(i)
		got := string(slot.Elem)
		assert.Equal(t, want, got, "error listening subscribed elements, wanted was %q but got %q", want, got)
	}
	assert.Nil(t, err, "channel remains open for future consumes")
}
func TestFanout_Unsubscribe(t *testing.T) {
	fo := flow.NewBufferedFanOut[string](10)
	// Adds one extra subscriber for test hardening.
	_, _ = fo.Subscribe() //nolint:dogsled
	_, unsubscribe := fo.Subscribe()

	err := unsubscribe()
	require.NoError(t, err)
	assert.Equal(t, 1, fo.ActiveSubscribers())

}

func TestFanout_Unsubscribe_NotFound(t *testing.T) {
	fo := flow.NewBufferedFanOut[string](10)
	_, cancel := fo.Subscribe()
	fo.Reset()
	err := cancel()
	assert.IsType(t, flow.ErrSubscriberNotFound, err, "wanted flow.ErrSubscriberNotFound got %T", err)
}

func TestFanout_Reset(t *testing.T) {
	fo := flow.NewBufferedFanOut[string](10)

	consume, _ := fo.Subscribe()

	fo.Add("dd")

	fo.Reset()

	assert.Equal(t, 0, fo.ActiveSubscribers(), "no subscribers expected after reset")
	// Check consume is closed after consumption
	consume()
	_, err := consume()
	assert.Equal(t, io.EOF, err)
}

func TestFanout_Add_NoActiveSubscriberDoesntBlock(t *testing.T) {
	maxBuffLen := 10
	fo := flow.NewBufferedFanOut[string](maxBuffLen)
	consume, _ := fo.Subscribe() // this subscriber will try to block the entire system

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
	for {
		s, err := consume()
		if err == io.EOF {
			break
		}
		assert.Contains(t, string(s.Elem), "dn",
			"The are elements in subs channel that needs to be discarded")
	}
}

func TestFanout_Status_Count_Aggregated(t *testing.T) {
	fo := flow.NewBufferedFanOut[int](10)

	_, _ = fo.Subscribe()
	_, _ = fo.Subscribe()

	fo.Add(1)
	fo.Add(2)

	want := flow.Status{
		"": 4,
	}
	assert.Equal(t, want, fo.Status())
}

func TestFanout_Status_Count(t *testing.T) {
	fo := flow.NewBufferedFanOut[int](10)

	_, _ = fo.SubscribeWith("a")
	consume, _ := fo.SubscribeWith("b")
	fo.Add(1)
	fo.Add(2)
	consume()
	want := flow.Status{
		"a": 2,
		"b": 1,
	}
	assert.Equal(t, want, fo.Status())
}

func TestFanout_Status_Unsubscribe(t *testing.T) {
	fo := flow.NewBufferedFanOut[int](10)

	_, cancel := fo.SubscribeWith("a")
	_, _ = fo.SubscribeWith("b")
	fo.Add(1)
	fo.Add(2)
	err := cancel()
	require.NoError(t, err)
	want := flow.Status{
		"b": 2,
	}
	assert.Equal(t, want, fo.Status())
}

func TestSubscribersStoreReuse(t *testing.T) {
	fo := flow.NewBufferedFanOut[int](10)

	fo.Subscribe() // 1
	fo.Subscribe() // 2

	// Cancelling the 3th subscription.
	_, cancel := fo.Subscribe() // 3
	cancel()

	fo.Subscribe() // 4 This subscriber should be allocated in same place of previous canceled subscription (3).

	assert.Equal(t, 3, fo.SubscribersLen(), "It should not reserve more subscriber slots, if we already have empty ones.")

}

func TestSubscribersStoreGrows(t *testing.T) {
	fo := flow.NewBufferedFanOut[int](10)

	fo.Subscribe()
	fo.Subscribe()
	fo.Subscribe()

	assert.Equal(t, 3, fo.SubscribersLen(), "Subscriber len should grow linearly")

}