package moment

import (
	"time"
)

type Current func() time.Time

func NewFakedCurrent(year int, month time.Month, day, hour, min, sec, nSec int, loc *time.Location) Current {
	return func() time.Time {
		return time.Date(year, month, day, hour, min, sec, nSec, loc)
	}
}
