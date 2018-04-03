package index

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type inMemoryIndex struct {
	entryMap map[string][]byte
}

func NewInMemoryIndex() Index {
	return &inMemoryIndex{
		entryMap: map[string][]byte{},
	}
}

func (ri *inMemoryIndex) Find(key interface{}, valuePtr interface{}) error {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshalling key %#v", key)
	}

	valueBytes, exists := ri.entryMap[string(keyBytes)]
	if !exists {
		return ErrNotFound
	}

	err = json.Unmarshal(valueBytes, valuePtr)
	if err != nil {
		return bosherr.WrapErrorf(err, "Unmarshaling value for key %#v", key)
	}

	return nil
}

func (ri *inMemoryIndex) Save(key interface{}, value interface{}) error {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshalling key %#v", key)
	}

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshalling value %#v", value)
	}

	ri.entryMap[string(keyBytes)] = valueBytes

	return nil
}
