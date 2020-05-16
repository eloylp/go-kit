package fanouttest

import (
	"github.com/eloylp/go-kit/flow/fanout"
	"github.com/eloylp/go-kit/moment"
	"strconv"
)

// PopulatedBufferedFanOut facilitates construction of an store
// by accepting the number of elements to be present, and
// store.BufferedStore needed constructor args.
func PopulatedBufferedFanOut(elems, maxBuffLen int, now moment.Now) *fanout.BufferedFanOut {
	fo := fanout.NewBufferedFanOut(maxBuffLen, now)
	Populate(fo, elems)
	return fo
}

func BufferedFanOut(maxBuffLen int, now moment.Now) *fanout.BufferedFanOut {
	fo := fanout.NewBufferedFanOut(maxBuffLen, now)
	return fo
}

func Populate(fo *fanout.BufferedFanOut, elems int) {
	for i := 0; i < elems; i++ {
		data := "d" + strconv.Itoa(i)
		elem := []byte(data)
		fo.AddElem(elem)
	}
}
