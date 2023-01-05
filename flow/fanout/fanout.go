package fanout

import (
	"io"
	"sync"
	"time"
)

// Slot represents an enqueueable element. Timestamp
// will allow consumers discard old messages. T will
// represent the user custom data.
type Slot[T any] struct {
	TimeStamp time.Time
	Elem      T
}

// ConsumerFunc represents the function type
// returned by the Subscribe() operation.
// It should be called by users in order
// to retrieve elements from the queue.
//
// In case there are no more elements, an io.EOF
// error will be returned.
type ConsumerFunc[T any] func() (*Slot[T], error)

// CancelFunc is the main way that a consumer can end its
// subscription. When called, the subscriber channel
// will be closed, so consumer will still try to
// consume all the remaining messages.
type CancelFunc func() error

// Status represents a report about how much
// elements are queued per consumer.
//
// As the Key, the provided subscriber UUID is used (See BufferedFanOut.SubscribeWith).
// If not provided, an empty string will be used, aggregating
// all counter values from all non customized consumers.
//
// As Value, the number of queued elements.
type Status map[string]int

// BufferedFanOut represent a fan-out pattern making
// use of channels.
//
// It Will send a copy of the element to multiple
// subscribers at the same time.
//
// This implements the needed locking mechanisms
// to be considered thread safe.
type BufferedFanOut[T any] struct {
	subscribers []*subscriber[T]
	maxBuffLen  int
	L           sync.RWMutex
}

type subscriber[T any] struct {
	ch   chan *Slot[T]
	UUID string
}

// NewBufferedFanOut needs buffer size for subscribers channels
// and function that must retrieve the current time for Slots
// timestamps.
func NewBufferedFanOut[T any](maxBuffLen int) *BufferedFanOut[T] {
	fo := &BufferedFanOut[T]{
		maxBuffLen: maxBuffLen,
	}
	return fo
}

// Add will send of elem to all subscribers channels.
// Take care about race conditions and the type of
// element that pass in. If you pass an integer, that
// will be copied to each subscriber. Problem comes if you
// pass referenced types like maps or slices or any other
// pointer type.
//
// If one of the subscribers channels is full, oldest data
// will be discarded.
func (fo *BufferedFanOut[T]) Add(elem T) {
	fo.L.Lock()
	defer fo.L.Unlock()
	sl := &Slot[T]{
		TimeStamp: time.Now(),
		Elem:      elem,
	}
	fo.publish(sl)
}

func (fo *BufferedFanOut[T]) publish(sl *Slot[T]) {
	for i := 0; i < len(fo.subscribers); i++ {
		if fo.subscribers[i] == nil {
			continue
		}
		if len(fo.subscribers[i].ch) == fo.maxBuffLen {
			<-fo.subscribers[i].ch // remove last Slot of subscriber channel
		}
		fo.subscribers[i].ch <- sl
	}
}

// ActiveSubscribers can tell us how many subscribers
// are registered and active in the present moment.
func (fo *BufferedFanOut[T]) ActiveSubscribers() int {
	fo.L.RLock()
	defer fo.L.RUnlock()
	var count int
	for i := 0; i < len(fo.subscribers); i++ {
		if fo.subscribers[i] != nil {
			count++
		}
	}
	return count
}

// SubscribersLen can tell us the size of the underlying
// subscriber storage. This will return both, active and
// non active slots.
func (fo *BufferedFanOut[T]) SubscribersLen() int {
	fo.L.RLock()
	defer fo.L.RUnlock()
	return len(fo.subscribers)
}

// Subscribe will return an output channel that will
// be filled when more data arrives this fanout. Also,
// a cancelFunc, for easily canceling the subscription.
//
// If you are not actively consuming this channel, but
// data continues arriving to the fanout, the oldest
// element will be dropped in favor of the new one.
func (fo *BufferedFanOut[T]) Subscribe() (ConsumerFunc[T], CancelFunc) { //nolint:gocritic
	return fo.SubscribeWith("")
}

func (fo *BufferedFanOut[T]) SubscribeWith(uuid string) (ConsumerFunc[T], CancelFunc) { //nolint:gocritic
	fo.L.Lock()
	defer fo.L.Unlock()
	ch := make(chan *Slot[T], fo.maxBuffLen)

	consumerFn := func() (*Slot[T], error) {
		slot, ok := <-ch
		if !ok {
			return slot, io.EOF
		}
		return slot, nil
	}

	subscriber := &subscriber[T]{ch, uuid}

	// Prefer reusing a free slot caused by a previous unsubscribe operation.
	// Try to not increase underlying array too much. This is O(n) worst case.
	for i := 0; i < len(fo.subscribers); i++ {
		if fo.subscribers[i] == nil {
			fo.subscribers[i] = subscriber
			return consumerFn, func() error {
				return fo.unsubscribe(i)
			}
		}
	}

	// Looks like we are full of subscribers. Time to append more ...
	fo.subscribers = append(fo.subscribers, subscriber)
	index := len(fo.subscribers) - 1
	return consumerFn, func() error {
		return fo.unsubscribe(index)
	}
}

func (fo *BufferedFanOut[T]) unsubscribe(index int) error {
	fo.L.Lock()
	defer fo.L.Unlock()

	if len(fo.subscribers) <= index {
		return ErrSubscriberNotFound
	}
	s := fo.subscribers[index]
	if s == nil {
		return ErrSubscriberNotFound
	}
	close(s.ch)
	fo.subscribers[index] = nil
	return nil
}

// Reset will clear all the data and
// subscribers, starting again. It will
// also close all the subscribers channels.
func (fo *BufferedFanOut[T]) Reset() {
	fo.L.Lock()
	defer fo.L.Unlock()
	for i := 0; i < len(fo.subscribers); i++ {
		if fo.subscribers[i] == nil {
			continue
		}
		close(fo.subscribers[i].ch)
	}
	fo.subscribers = nil
}

// Status will return a Status type with
// the list of all subscribers and they
// pending elements.
func (fo *BufferedFanOut[T]) Status() Status {
	fo.L.RLock()
	defer fo.L.RUnlock()

	status := make(Status, len(fo.subscribers))
	for i := 0; i < len(fo.subscribers); i++ {
		if fo.subscribers[i] == nil {
			continue
		}
		if fo.subscribers[i].UUID == "" {
			status[""] += len(fo.subscribers[i].ch)
			continue
		}
		status[fo.subscribers[i].UUID] = len(fo.subscribers[i].ch)
	}
	return status
}
