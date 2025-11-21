package generic_test

import (
	. "github.com/cloudfoundry/cli/generic"
	"testing"
)

func BenchmarkMerge_SmallMaps(b *testing.B) {
	map1 := NewMap(map[interface{}]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	map2 := NewMap(map[interface{}]interface{}{
		"key3": "value3",
		"key4": "value4",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Merge(map1, map2)
	}
}

func BenchmarkMerge_LargeMaps(b *testing.B) {
	map1Data := make(map[interface{}]interface{})
	map2Data := make(map[interface{}]interface{})

	for i := 0; i < 100; i++ {
		map1Data[i] = i
		map2Data[i+100] = i + 100
	}

	map1 := NewMap(map1Data)
	map2 := NewMap(map2Data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Merge(map1, map2)
	}
}

func BenchmarkDeepMerge_TwoMaps(b *testing.B) {
	map1 := NewMap(map[interface{}]interface{}{
		"key1": "value1",
	})
	map2 := NewMap(map[interface{}]interface{}{
		"key2": "value2",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeepMerge(map1, map2)
	}
}

func BenchmarkDeepMerge_MultipleMaps(b *testing.B) {
	maps := make([]Map, 10)
	for i := 0; i < 10; i++ {
		maps[i] = NewMap(map[interface{}]interface{}{
			i: i,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeepMerge(maps...)
	}
}

func BenchmarkDeepMerge_NestedMaps(b *testing.B) {
	map1 := NewMap(map[interface{}]interface{}{
		"outer": map[interface{}]interface{}{
			"inner1": "value1",
			"inner2": "value2",
		},
	})
	map2 := NewMap(map[interface{}]interface{}{
		"outer": map[interface{}]interface{}{
			"inner3": "value3",
			"inner4": "value4",
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeepMerge(map1, map2)
	}
}

func BenchmarkReduce_SmallCollection(b *testing.B) {
	collections := []Map{
		NewMap(map[interface{}]interface{}{"key1": "value1"}),
		NewMap(map[interface{}]interface{}{"key2": "value2"}),
		NewMap(map[interface{}]interface{}{"key3": "value3"}),
	}

	reducer := func(key, val interface{}, reducedVal Map) Map {
		reducedVal.Set(key, val)
		return reducedVal
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reduce(collections, NewMap(), reducer)
	}
}

func BenchmarkReduce_LargeCollection(b *testing.B) {
	collections := make([]Map, 100)
	for i := 0; i < 100; i++ {
		collections[i] = NewMap(map[interface{}]interface{}{i: i})
	}

	reducer := func(key, val interface{}, reducedVal Map) Map {
		reducedVal.Set(key, val)
		return reducedVal
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reduce(collections, NewMap(), reducer)
	}
}
