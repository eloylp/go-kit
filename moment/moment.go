package moment

import (
	"time"
)

type Now func() time.Time

func NewFakedNow(year int, month time.Month, day, hour, min, sec, nSec int, loc *time.Location) Now {
	return func() time.Time {
		return time.Date(year, month, day, hour, min, sec, nSec, loc)
	}
}
