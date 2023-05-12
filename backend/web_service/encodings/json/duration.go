package json

import (
	"errors"
	"strconv"
)

type JsonDuration struct {
	dur  int
	unit string
}

func (d JsonDuration) MarshalJSON() ([]byte, error) {
	if d.unit != "sec" && d.unit != "min" && d.unit != "hour" {
		return nil, errors.New("invalid duration unit " + d.unit)
	}
	return []byte(strconv.Itoa(d.dur) + " " + d.unit), nil
}

func (d *JsonDuration) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var dur int
	var unit string
	for i, c := range b {
		if c >= '0' && c <= '9' {
			dur = dur*10 + int(c)
		}
		if c >= 'a' && c <= 'z' {
			unit = string(b[i:])
			break
		}
	}
	if unit != "sec" && unit != "min" && unit != "hour" {
		return errors.New("invalid duration unit " + unit)
	}

	*d = JsonDuration{
		dur:  dur,
		unit: unit,
	}
	return nil
}

func NewJsonDuration(dur int, unit string) *JsonDuration {
	return &JsonDuration{
		dur:  dur,
		unit: unit,
	}
}
