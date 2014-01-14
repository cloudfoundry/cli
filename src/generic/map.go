package generic

import "fmt"

type Iterator func (key, val interface{})

func Merge(collection, otherCollection Map) interface{} {
	mergedMap := NewMap()

	iterator := func(key, value interface{}) () {
		mergedMap.Set(key, value)
	}

	Each(collection, iterator)
	Each(otherCollection, iterator)

	return mergedMap
}

func Each(collection Map, cb Iterator) {
	for _, key := range collection.Keys() {
		cb(key, collection.Get(key))
	}
}

type Map interface {
	IsEmpty() bool
	Count() int
	Keys() []interface{}
	Has(key interface{}) bool
	IsNil(key interface{}) bool
	NotNil(key interface{}) bool
	Get(key interface{}) interface{}
	Set(key interface{}, value interface{})
	Delete(key interface{})
}

type ConcreteMap map[interface{}]interface{}

func (data *ConcreteMap) String() string {
	return fmt.Sprintf("% v",*data)
}

func (data *ConcreteMap) IsEmpty() bool {
	return data.Count() == 0
}

func (data *ConcreteMap) Count() int {
	return len(*data)
}

func (data *ConcreteMap) Has(key interface{}) bool {
	_, ok := (*data)[key]
	return ok
}

func (data *ConcreteMap) IsNil(key interface{}) bool {
	maybe, ok := (*data)[key]
	return ok && maybe == nil
}

func (data *ConcreteMap) NotNil(key interface{}) bool {
	maybe, ok := (*data)[key]
	return ok && maybe != nil
}

func (data *ConcreteMap) Keys() (keys []interface{}) {
	keys = make([]interface{}, 0, data.Count())
	for key := range *data {
		keys = append(keys, key)
	}

	return
}

func (data *ConcreteMap) Get(key interface{}) interface{} {
	return (*data)[key]
}

func (data *ConcreteMap) Set(key, value interface{}) {
	(*data)[key] = value
}

func (data *ConcreteMap) Delete(key interface{}) {
	delete(*data, key)
}

func newEmptyMap() Map {
	return &ConcreteMap{}
}

func NewMap(data ...interface {}) Map {
	if len(data) == 0 {
		return newEmptyMap()
	} else if len(data) > 1 {
		panic("NewMap called with more than one argument")
	}

	switch data := data[0].(type){
	case Map:
		return data
	case map[string]string:
		stringMap := newEmptyMap()
		for key, val := range data {
			stringMap.Set(key, val)
		}
		return stringMap
	case map[interface {}]interface{}:
		mapp := ConcreteMap(data)
		return &mapp
	}
	panic("NewMap called with unexpected argument")
}
