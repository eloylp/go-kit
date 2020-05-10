package fanout_test

import (
	"testing"

	"github.com/eloylp/go-kit/flow/fanout/fanouttest"
)

func BufferedFanOutAddElem(b *testing.B, maxItems, subscriberBuffSize, subscribers, messagesSize int) {
	b.ReportAllocs()
	fo := fanouttest.PopulatedBufferedFanOut(b, 0, maxItems, subscriberBuffSize)
	for i := 0; i < subscribers; i++ {
		fo.Subscribe()
	}
	b.ResetTimer()
	data := make([]byte, messagesSize)
	for n := 0; n < b.N; n++ {
		err := fo.AddElem(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBufferedFanOut_AddItem_10_10_100_2(b *testing.B) {
	BufferedFanOutAddElem(b, 10, 10, 100, 2)
}

func BenchmarkBufferedFanOut_AddItem_10_10_1000_2(b *testing.B) {
	BufferedFanOutAddElem(b, 10, 10, 1000, 2)
}

func BenchmarkBufferedFanOut_AddItem_100_100_10000_200(b *testing.B) {
	BufferedFanOutAddElem(b, 100, 100, 10000, 200)
}

func BenchmarkBufferedFanOut_AddItem_100_100_10000_2(b *testing.B) {
	BufferedFanOutAddElem(b, 100, 100, 10000, 2)
}

func BenchmarkBufferedFanOut_AddItem_100_100_100000_200(b *testing.B) {
	BufferedFanOutAddElem(b, 100, 100, 100000, 200)
}

func BenchmarkBufferedFanOut_AddItem_100_100_1000000_200(b *testing.B) {
	BufferedFanOutAddElem(b, 100, 100, 1000000, 200)
}
