package fanout

import (
	"sync"
	"time"

	guuid "github.com/google/uuid"
)

type BufferedFanOut struct {
	slots              []*Slot
	subscribers        []subscriber
	maxElems           int
	subscriberBuffSize int
	L                  sync.RWMutex
}

type subscriber struct {
	ch   chan *Slot
	UUID string
}

type Slot struct {
	TimeStamp time.Time
	Elem      interface{}
}

func (fo *BufferedFanOut) Length() int {
	fo.L.RLock()
	defer fo.L.RUnlock()
	return len(fo.slots)
}

func (fo *BufferedFanOut) Subscribers() int {
	fo.L.RLock()
	defer fo.L.RUnlock()
	return len(fo.subscribers)
}

func NewBufferedFanOut(maxSlots, subscriberBuffSize int) *BufferedFanOut {
	return &BufferedFanOut{maxElems: maxSlots, subscriberBuffSize: subscriberBuffSize}
}

func (fo *BufferedFanOut) AddElem(elem interface{}) error {
	fo.L.Lock()
	defer fo.L.Unlock()
	s := &Slot{
		TimeStamp: time.Now(),
		Elem:      elem,
	}
	fo.addSlot(s)
	fo.gcItems()
	fo.publish(s)
	return nil
}

func (fo *BufferedFanOut) addSlot(s *Slot) {
	fo.slots = append(fo.slots, s)
}

func (fo *BufferedFanOut) gcItems() {
	l := len(fo.slots)
	if l > fo.maxElems {
		i := make([]*Slot, fo.maxElems)
		e := fo.slots[1:l]
		copy(i, e)
		fo.slots = i
	}
}

// Subscribers that doesnt consume elements will begin to
// loose old ones.
func (fo *BufferedFanOut) publish(sl *Slot) {
	for _, s := range fo.subscribers {
		if len(s.ch) == fo.subscriberBuffSize {
			<-s.ch // remove last Slot of subscriber channel
		}
		s.ch <- sl
	}
}

func (fo *BufferedFanOut) Subscribe() (<-chan *Slot, string) { //nolint:gocritic
	fo.L.Lock()
	defer fo.L.Unlock()
	ch := make(chan *Slot, fo.subscriberBuffSize)
	uuid := guuid.New().String()
	fo.subscribers = append(fo.subscribers, subscriber{ch, uuid})
	for _, sl := range fo.slots {
		if sl != nil {
			ch <- sl
		}
	}
	return ch, uuid
}

func (fo *BufferedFanOut) Unsubscribe(uuid string) error {
	if !fo.exists(uuid) {
		return ErrSubscriberNotFound
	}
	fo.L.Lock()
	defer fo.L.Unlock()
	ec := fo.estimateCapacity()
	newSubs := make([]subscriber, 0, ec)
	for _, s := range fo.subscribers {
		if s.UUID == uuid {
			close(s.ch)
		} else {
			newSubs = append(newSubs, s)
		}
	}
	fo.subscribers = newSubs
	return nil
}

func (fo *BufferedFanOut) exists(uuid string) bool {
	fo.L.RLock()
	defer fo.L.RUnlock()
	for _, s := range fo.subscribers {
		if s.UUID == uuid {
			return true
		}
	}
	return false
}

func (fo *BufferedFanOut) estimateCapacity() int {
	cl := len(fo.subscribers)
	ec := 0
	if cl > 0 {
		ec = cl - 1
	}
	return ec
}

func (fo *BufferedFanOut) Reset() {
	fo.L.Lock()
	defer fo.L.Unlock()
	for _, s := range fo.subscribers {
		close(s.ch)
	}
	fo.subscribers = nil
	fo.slots = nil
}
