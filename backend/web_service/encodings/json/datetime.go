package json

import (
	"errors"
	"time"
)

var defaultFormat string

type JsonDateTime struct {
	t   time.Time
	fmt string
}

func (dt JsonDateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + dt.t.Format(dt.fmt) + `"`), nil
}

func (dt *JsonDateTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	if (dt == nil || len(dt.fmt) == 0) && len(defaultFormat) == 0 {
		return errors.New("undefined datetime format. Either use SetDefaultFormat() or create a JsonDateTime value with a format using NewJsonDateTime()")
	}
	var t time.Time
	var err error
	var fmt string
	if dt != nil && len(dt.fmt) > 0 {
		t, err = time.Parse(dt.fmt, string(b[1:len(b)-1]))
		fmt = dt.fmt
	} else {
		t, err = time.Parse(defaultFormat, string(b[1:len(b)-1]))
		fmt = defaultFormat
	}

	if err != nil {
		return err
	}

	dt = &JsonDateTime{
		t:   t,
		fmt: fmt,
	}
	return nil
}

func (dt JsonDateTime) Compare(odt JsonDateTime) int {
	t1 := time.Time(dt.t)
	t2 := time.Time(odt.t)
	if t1.Before(t2) {
		return -1
	}
	if t1.After(t2) {
		return 1
	}
	return 0
}

func SetDefaultFormat(fmt string) {
	defaultFormat = fmt
}

func NewJsonDateTime(t time.Time, fmt string) *JsonDateTime {
	return &JsonDateTime{
		t:   t,
		fmt: fmt,
	}
}
