package maker

import (
	"fmt"
	"strings"
)

type Overrides map[string]interface{}

func (params Overrides) canonicalKey(key interface{}) string {
	return strings.ToLower(key.(string))
}

func (params Overrides) Has(key interface{}) bool {
	_, ok := params[params.canonicalKey(key)]
	return ok
}

func (params Overrides) Get(key interface{}) interface{} {
	return params[params.canonicalKey(key)]
}

func guidGenerator(prefix string) func() string {
	count := 0
	return func() string {
		count++
		return fmt.Sprintf("%s-guid-%d", prefix, count)
	}
}
