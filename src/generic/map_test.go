package generic

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestDeepMerge(t *testing.T) {
	map1 := NewMap(map[interface{}]interface{}{
		"key1": "val1",
		"key2": "val2",
		"nest1": map[interface{}]interface{}{
			"nestKey1":"nest1Val1",
			"nestKey2":"nest1Val2",
		},
		"nest2": []interface{}{
			"nest2Val1",
		},
	})

	map2 := NewMap(map[interface{}]interface{}{
		"key1": "newVal1",
		"nest1": map[interface{}]interface{}{
			"nestKey1":"newNest1Val1",
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
	fmt.Printf("Merged Map:\n%s\n",mergedMap)
	fmt.Printf("Expected Map:\n%s\n",expectedMap)
	assert.Equal(t, mergedMap, expectedMap)
}
