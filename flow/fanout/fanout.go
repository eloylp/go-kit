package fanout

import (
	"sync"
	"time"

	guuid "github.com/google/uuid"
)

type BufferedFanOut struct {
	subscribers        []subscriber
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

func (fo *BufferedFanOut) Subscribers() int {
	fo.L.RLock()
	defer fo.L.RUnlock()
	return len(fo.subscribers)
}

func NewBufferedFanOut(subscriberBuffSize int) *BufferedFanOut {
	fo := &BufferedFanOut{
		subscriberBuffSize: subscriberBuffSize}
	return fo
}

// AddElem publish
//Subscribers that doesnt consume elements will begin to
// loose old ones.
func (fo *BufferedFanOut) AddElem(elem interface{}) {
	fo.L.Lock()
	defer fo.L.Unlock()
	sl := &Slot{
		TimeStamp: time.Now(),
		Elem:      elem,
	}
	fo.publish(sl)
}

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
	return ch, uuid
}

func (fo *BufferedFanOut) Unsubscribe(uuid string) error {
	fo.L.Lock()
	defer fo.L.Unlock()
	if !fo.exists(uuid) {
		return ErrSubscriberNotFound
	}
	newSubs := make([]subscriber, 0, len(fo.subscribers))
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
	for _, s := range fo.subscribers {
		if s.UUID == uuid {
			return true
		}
	}
	return false
}

func (fo *BufferedFanOut) Reset() {
	fo.L.Lock()
	defer fo.L.Unlock()
	for _, s := range fo.subscribers {
		close(s.ch)
	}
	fo.subscribers = nil
}
