package fanout_test

import (
	"testing"

	"github.com/eloylp/go-kit/flow/fanout/fanouttest"
)

func BufferedFanOutAddElem(b *testing.B, subscribers, subscriberBuffSize, messagesSize int) {
	b.ReportAllocs()
	fo := fanouttest.BufferedFanOut(subscriberBuffSize)
	for i := 0; i < subscribers; i++ {
		fo.Subscribe()
	}
	b.ResetTimer()
	data := make([]byte, messagesSize)
	for n := 0; n < b.N; n++ {
		fo.AddElem(data)
	}
}

func BenchmarkBufferedFanOut_AddItem_10_50_10000(b *testing.B) {
	BufferedFanOutAddElem(b, 10, 50, 10000)
}

func BenchmarkBufferedFanOut_AddItem_100_50_10000(b *testing.B) {
	BufferedFanOutAddElem(b, 100, 50, 10000)
}

func BenchmarkBufferedFanOut_AddItem_1000_50_10000(b *testing.B) {
	BufferedFanOutAddElem(b, 1000, 50, 10000)
}

func BenchmarkBufferedFanOut_AddItem_10000_50_10000(b *testing.B) {
	BufferedFanOutAddElem(b, 10000, 50, 10000)
}

func BenchmarkBufferedFanOut_AddItem_100000_50_10000(b *testing.B) {
	BufferedFanOutAddElem(b, 100000, 50, 10000)
}

func BenchmarkBufferedFanOut_AddItem_1000000_50_10000(b *testing.B) {
	BufferedFanOutAddElem(b, 1000000, 50, 10000)
}
