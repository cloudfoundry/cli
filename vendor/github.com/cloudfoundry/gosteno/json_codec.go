package gosteno

import (
	"encoding/json"
)

type JsonCodec struct {
}

func NewJsonCodec() Codec {
	return new(JsonCodec)
}

func (j *JsonCodec) EncodeRecord(record *Record) ([]byte, error) {
	b, err := json.Marshal(record)
	if err != nil {
		return json.Marshal(map[string]string{"error": err.Error()})
	}

	return b, err
}
