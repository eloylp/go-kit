package fanouttest

import (
	"strconv"

	"go.eloylp.dev/kit/flow/fanout"
	"go.eloylp.dev/kit/moment"
)

// BufferedFanOut is here for hiding constructor in tests.
// Will return an empty fanout instance.
func BufferedFanOut(maxBuffLen int, now moment.Now) *fanout.BufferedFanOut {
	fo := fanout.NewBufferedFanOut(maxBuffLen, now)
	return fo
}

// Populate must be used in tests to fill a BufferedFanOut
// with elements. Its target is to preserve testing more
// consistent.
func Populate(fo *fanout.BufferedFanOut, elems int) {
	for i := 0; i < elems; i++ {
		data := "d" + strconv.Itoa(i)
		elem := []byte(data)
		fo.Add(elem)
	}
}
