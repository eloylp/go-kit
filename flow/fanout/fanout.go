package fanout

import (
	"github.com/eloylp/go-kit/moment"
	"sync"
	"time"

	guuid "github.com/google/uuid"
)

type Status map[string]int
type CancelFunc func() error

type subscriber struct {
	ch   chan *Slot
	UUID string
}

type Slot struct {
	TimeStamp time.Time
	Elem      interface{}
}

type BufferedFanOut struct {
	subscribers []subscriber
	maxBuffLen  int
	L           sync.RWMutex
	Now         moment.Now
}

func NewBufferedFanOut(subscriberBuffSize int, now moment.Now) *BufferedFanOut {
	fo := &BufferedFanOut{
		maxBuffLen: subscriberBuffSize,
		Now:        now,
	}
	return fo
}

// AddElem publish
//Subscribers that doesnt consume elements will begin to
// loose old ones.
func (fo *BufferedFanOut) AddElem(elem interface{}) {
	fo.L.Lock()
	defer fo.L.Unlock()
	sl := &Slot{
		TimeStamp: fo.Now(),
		Elem:      elem,
	}
	fo.publish(sl)
}

func (fo *BufferedFanOut) Subscribers() int {
	fo.L.RLock()
	defer fo.L.RUnlock()
	return len(fo.subscribers)
}

func (fo *BufferedFanOut) publish(sl *Slot) {
	for _, s := range fo.subscribers {
		if len(s.ch) == fo.maxBuffLen {
			<-s.ch // remove last Slot of subscriber channel
		}
		s.ch <- sl
	}
}

func (fo *BufferedFanOut) Subscribe() (<-chan *Slot, string, CancelFunc) { //nolint:gocritic
	fo.L.Lock()
	defer fo.L.Unlock()
	ch := make(chan *Slot, fo.maxBuffLen)
	uuid := guuid.New().String()
	fo.subscribers = append(fo.subscribers, subscriber{ch, uuid})
	return ch, uuid, func() error {
		return fo.Unsubscribe(uuid)
	}
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

func (fo *BufferedFanOut) Status() Status {
	fo.L.RLock()
	defer fo.L.RUnlock()
	status := make(Status, len(fo.subscribers))
	for _, s := range fo.subscribers {
		status[s.UUID] = len(s.ch)
	}
	return status
}
