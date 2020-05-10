package fanouttest

import (
	"strconv"
	"testing"

	"github.com/eloylp/go-kit/flow/fanout"
)

// PopulatedBufferedFanOut facilitates construction of an store
// by accepting the number of elements to be present, and
// store.BufferedStore needed constructor args.
func PopulatedBufferedFanOut(t testing.TB, elems, maxElems, subscriberBuffSize int) *fanout.BufferedFanOut {
	s := fanout.NewBufferedFanOut(maxElems, subscriberBuffSize)
	for i := 0; i < elems; i++ {
		data := "d" + strconv.Itoa(i)
		elem := []byte(data)
		if err := s.AddElem(elem); err != nil {
			t.Fatal(err)
		}
	}
	return s
}
