package trips

import (
	"time"
)

func createJsonTime(t *time.Time) Datetime {
	return Datetime{
		Year:  t.Year(),
		Month: int(t.Month()),
		Day:   t.Day(),
		Hour:  t.Hour(),
		Min:   t.Minute(),
	}
}
