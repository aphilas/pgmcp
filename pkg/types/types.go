package types

import "encoding/json"

func Ptr[T any](v T) *T {
	return &v
}

func NewRawJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(b)
}
