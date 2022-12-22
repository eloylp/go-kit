package fanout_test

import (
	"testing"
	"time"

	"go.eloylp.dev/kit/flow/fanout"
)

func BufferedFanOutAddElem(b *testing.B, subscribers, maxBuffLen, msgLen int) {
	b.ReportAllocs()

	fo := fanout.NewBufferedFanOut[[]byte](maxBuffLen, time.Now)
	for i := 0; i < subscribers; i++ {
		fo.Subscribe()
	}
	data := make([]byte, msgLen)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		fo.Add(data)
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
