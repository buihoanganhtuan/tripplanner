package trips

import (
	"time"

	constants "github.com/buihoanganhtuan/tripplanner/backend/web_service/_constants"
)

func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse(constants.DatetimeFormat, s)
	if err != nil {
		return time.Time{}, StatusError{
			Status: ParseError,
			Err:    err,
		}
	}
	return t, nil
}
