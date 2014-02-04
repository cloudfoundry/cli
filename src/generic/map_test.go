package generic

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestDeepMerge", func() {

			map1 := NewMap(map[interface{}]interface{}{
				"key1": "val1",
				"key2": "val2",
				"nest1": map[interface{}]interface{}{
					"nestKey1": "nest1Val1",
					"nestKey2": "nest1Val2",
				},
				"nest2": []interface{}{
					"nest2Val1",
				},
			})

			map2 := NewMap(map[interface{}]interface{}{
				"key1": "newVal1",
				"nest1": map[interface{}]interface{}{
					"nestKey1": "newNest1Val1",
				},
				"nest2": []interface{}{
					"something",
				},
			})

			expectedMap := NewMap(map[interface{}]interface{}{
				"key1": "newVal1",
				"key2": "val2",
				"nest1": NewMap(map[interface{}]interface{}{
					"nestKey1": "newNest1Val1",
					"nestKey2": "nest1Val2",
				}),
				"nest2": []interface{}{
					"nest2Val1",
					"something",
				},
			})

			mergedMap := DeepMerge(map1, map2)
			fmt.Printf("Merged Map:\n%s\n", mergedMap)
			fmt.Printf("Expected Map:\n%s\n", expectedMap)
			assert.Equal(mr.T(), mergedMap, expectedMap)
		})
	})
}
