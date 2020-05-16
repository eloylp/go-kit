package fanouttest

import (
	"github.com/eloylp/go-kit/flow/fanout"
	"strconv"
)

// PopulatedBufferedFanOut facilitates construction of an store
// by accepting the number of elements to be present, and
// store.BufferedStore needed constructor args.
func PopulatedBufferedFanOut(elems, subscriberBuffSize int) *fanout.BufferedFanOut {
	fo := fanout.NewBufferedFanOut(subscriberBuffSize)
	Populate(fo, elems)
	return fo
}

func BufferedFanOut(subscriberBuffSize int) *fanout.BufferedFanOut {
	fo := fanout.NewBufferedFanOut(subscriberBuffSize)
	return fo
}

func Populate(fo *fanout.BufferedFanOut, elems int) {
	for i := 0; i < elems; i++ {
		data := "d" + strconv.Itoa(i)
		elem := []byte(data)
		fo.AddElem(elem)
	}
}
