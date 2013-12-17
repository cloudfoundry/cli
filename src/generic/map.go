package generic

import "strings"

type Map map[interface{}]interface{}

func canonicalKey(key interface{}) string {
	return strings.ToLower(key.(string))
}

func (data Map) IsEmpty() bool {
	return len(data) == 0
}

func (data Map) Has(key interface{}) bool {
	_, ok := data[canonicalKey(key)]
	return ok
}

func (data Map) Get(key interface{}) interface{} {
	return data[canonicalKey(key)]
}

func (data Map) Set(key interface{}, value interface{}) {
	data[canonicalKey(key)] = value
}

func NewEmptyMap() Map {
	return Map{}
}

func NewMap(data interface {}) Map {
	switch data := data.(type){
	case map[interface {}]interface{}:
		return Map(data)
	default:
		return Map{}
	}
}
