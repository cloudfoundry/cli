package generic

type Iterator func (key, val interface{})

func Merge(collection, otherCollection Map) interface{} {
	mergedMap := NewEmptyMap()

	iterator := func(key interface{}, value interface{}) () {
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

func (data *ConcreteMap) Set(key interface{}, value interface{}) {
	(*data)[key] = value
}

func (data *ConcreteMap) Delete(key interface{}) {
	delete(*data, key)
}

func NewEmptyMap() Map {
	return &ConcreteMap{}
}

func NewMap(data interface {}) Map {
	switch data := data.(type){
	case Map:
		return data
	case map[string]string:
		stringMap := NewEmptyMap()
		for key, val := range data {
			stringMap.Set(key, val)
		}
		return stringMap
	case map[interface {}]interface{}:
		mapp := ConcreteMap(data)
		return &mapp
	default:
		return &ConcreteMap{}
	}
}
