package moment

import (
	"time"
)

// Now represents a function that returns
// current time. The target is to decouple
// time gathering operation from production
// code and in the manner, facilitate time
// based tests. This type can be used in
// struts, even as a function parameter.
type Now func() time.Time

// NewFakedNow returns a faked Now object
// ready to be used in tests.
func NewFakedNow(year int, month time.Month, day, hour, min, sec, nSec int, loc *time.Location) Now {
	return func() time.Time {
		return time.Date(year, month, day, hour, min, sec, nSec, loc)
	}
}
