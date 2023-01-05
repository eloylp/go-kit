package flow

import (
	"io"
	"sync"
	"time"
)

// Slot represents an enqueueable element. Timestamp
// will allow consumers discard old elements. T will
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
// In case there are no more elements the function
// will block.
//
// In case the subscription is cancelled, an io.EOF
// error will be returned.
type ConsumerFunc[T any] func() (*Slot[T], error)

// CancelFunc is the way that a consumer can terminate
// its subscription. When called, the subscriber
// channel will be closed, so consumer could still try to
// consume all the remaining elements.
type CancelFunc func() error

// Status represents a report about how much
// elements are queued per consumer.
//
// As the Key, the user provided subscriber UUID is used.
// If not provided, an empty string will be used, aggregating
// all counter values from all non customized consumers.
//
// As the Value, the number of queued elements.
type Status map[string]int

// Fanout represents a fan-out in-memory pattern
// with a configurable buffer.
//
// It Will send a copy of the element to multiple
// subscribers at the same time. In case a consumer
// gets stalled, older elements will be discarded
// in favour of the new arrived ones.
//
// This implements all the needed locking mechanisms,
// so it can be considered thread safe.
type Fanout[T any] struct {
	subscribers []*subscriber[T]
	maxBuffLen  int
	l           sync.RWMutex
}

type subscriber[T any] struct {
	ch   chan *Slot[T]
	uuid string
}

// NewFanout is the constructor for BufferedFanOut.
func NewFanout[T any](maxBuffLen int) *Fanout[T] {
	fo := &Fanout[T]{
		maxBuffLen: maxBuffLen,
	}
	return fo
}

// Add will send an elem to all subscribers channels.
// If one of the subscribers channels is full, oldest data
// will be discarded.
func (fo *Fanout[T]) Add(elem T) {
	fo.l.Lock()
	defer fo.l.Unlock()
	sl := &Slot[T]{
		TimeStamp: time.Now(),
		Elem:      elem,
	}
	fo.publish(sl)
}

func (fo *Fanout[T]) publish(sl *Slot[T]) {
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

// ActiveSubscribers will tell us how many subscribers
// are registered and active in the present moment.
func (fo *Fanout[T]) ActiveSubscribers() int {
	fo.l.RLock()
	defer fo.l.RUnlock()

	var count int
	for i := 0; i < len(fo.subscribers); i++ {
		if fo.subscribers[i] != nil {
			count++
		}
	}
	return count
}

// SubscribersLen returns the size of the underlying
// subscriber storage. This will return both, active and
// non active (free) subscriber slots.
func (fo *Fanout[T]) SubscribersLen() int {
	fo.l.RLock()
	defer fo.l.RUnlock()
	return len(fo.subscribers)
}

// Subscribe will return two functions, a ConsumerFunc and
// a CancelFunc (see types definitions).
//
// The ConsumerFunc will be used for pulling data from the
// fanout system, whereas the CancelFunc should always be
// used when the consumer its not interested on continuing
// its activity.
//
// Its IMPORTANT to not forget to use the CancelFunc to cancel
// the consumer activity. If not, resources could be leaked. A
// good practice here would be to always call the CancelFunc
// in a defer statement.
func (fo *Fanout[T]) Subscribe() (ConsumerFunc[T], CancelFunc) { //nolint:gocritic
	return fo.SubscribeWith("")
}

// SubscribeWith is same as Subscribe, but it allows to
// customize the subscriber with an UUID. Which might
// unequivocally identify a subscriber or a group of them
// in the system.
func (fo *Fanout[T]) SubscribeWith(uuid string) (ConsumerFunc[T], CancelFunc) { //nolint:gocritic
	fo.l.Lock()
	defer fo.l.Unlock()
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

func (fo *Fanout[T]) unsubscribe(index int) error {
	fo.l.Lock()
	defer fo.l.Unlock()

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
// also close all the subscribers internally.
//
// After calling Reset(), Subscribers can still
// consume all the remaining elements.
func (fo *Fanout[T]) Reset() {
	fo.l.Lock()
	defer fo.l.Unlock()
	for i := 0; i < len(fo.subscribers); i++ {
		if fo.subscribers[i] == nil {
			continue
		}
		close(fo.subscribers[i].ch)
	}
	fo.subscribers = nil
}

// Status will return a Status type with
// the list of all subscribers and their
// pending elements.
func (fo *Fanout[T]) Status() Status {
	fo.l.RLock()
	defer fo.l.RUnlock()

	status := make(Status, len(fo.subscribers))
	for i := 0; i < len(fo.subscribers); i++ {
		if fo.subscribers[i] == nil {
			continue
		}		
		status[fo.subscribers[i].uuid] += len(fo.subscribers[i].ch)
	}
	return status
}
