package registry

type registry map[string][]byte

type Registry interface {
	Save(string, []byte) bool
	Get(string) ([]byte, bool)
	Delete(string)
}

func NewRegistry() Registry {
	return make(registry)
}

func (r registry) Save(key string, value []byte) bool {
	_, exists := r[key]
	r[key] = value

	return exists
}

func (r registry) Get(key string) ([]byte, bool) {
	value, exists := r[key]

	return value, exists
}

func (r registry) Delete(key string) {
	delete(r, key)
}
