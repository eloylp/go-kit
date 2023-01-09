package moment

import (
	"testing"
	"time"
)

// layout is the date format that the time functions of this
// package will support.
const layout = "2006-01-02 15:04:05"

// NowFunc represents a function that returns current time.
// The target is to decouple time gathering operation
// from production code and in the manner, facilitate time
// based tests. This type can be used in structs, even as a function parameter.
type NowFunc func() time.Time

// NewFakedNow returns a faked Now type
// ready to be used in tests.
func NewFakedNow(t *testing.T, date string) NowFunc {
	d, err := time.Parse(layout, date)
	if err != nil {
		t.Fatal(err)
	}
	return func() time.Time {
		return d
	}
}

// NewFakedNowWithLoc returns a faked Now type
// ready to be used in tests. Needs a location as second
// parameter.
func NewFakedNowWithLoc(t *testing.T, date string, loc *time.Location) NowFunc {
	d, err := time.ParseInLocation(layout, date, loc)
	if err != nil {
		t.Fatal(err)
	}
	return func() time.Time {
		return d
	}
}
