package maker

import (
	"fmt"
)

type Overrides map[string]interface{}

func (params Overrides) Has(key interface{}) bool {
	_, ok := params[key.(string)]
	return ok
}

func guidGenerator(prefix string) func () string {
	count := 0
	return func () string {
		count++
		return fmt.Sprintf("%s-guid-%d", prefix, count)
	}
}
