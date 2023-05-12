package json

import (
	gjson "encoding/json"
)

// Optional[T] represent an optional JSON field. The advantage of using Optional[T] instead of *T
// for optional fields is that missing, null, and zero values are handled in the same way

type Optional[T any] struct {
	defined bool
	value   T
}

func (o Optional[T]) MarshalJSON() ([]byte, error) {
	return gjson.Marshal(o.value)
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.defined = true
	return gjson.Unmarshal(data, &o.value)
}
