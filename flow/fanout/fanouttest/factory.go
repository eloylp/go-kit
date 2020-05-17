package fanouttest

import (
	"github.com/eloylp/go-kit/flow/fanout"
	"github.com/eloylp/go-kit/moment"
	"strconv"
)

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
