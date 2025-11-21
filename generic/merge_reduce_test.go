package generic_test

import (
	. "github.com/cloudfoundry/cli/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Merge and Reduce", func() {
	Describe("Merge", func() {
		It("merges two maps together", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"key1": "value1",
				"key2": "value2",
			})

			map2 := NewMap(map[interface{}]interface{}{
				"key3": "value3",
				"key4": "value4",
			})

			result := Merge(map1, map2)

			Expect(result.Get("key1")).To(Equal("value1"))
			Expect(result.Get("key2")).To(Equal("value2"))
			Expect(result.Get("key3")).To(Equal("value3"))
			Expect(result.Get("key4")).To(Equal("value4"))
		})

		It("overwrites values from first map with second map", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"key1": "original-value",
				"key2": "value2",
			})

			map2 := NewMap(map[interface{}]interface{}{
				"key1": "overwritten-value",
			})

			result := Merge(map1, map2)

			Expect(result.Get("key1")).To(Equal("overwritten-value"))
			Expect(result.Get("key2")).To(Equal("value2"))
		})

		It("handles empty first map", func() {
			map1 := NewMap()
			map2 := NewMap(map[interface{}]interface{}{
				"key1": "value1",
			})

			result := Merge(map1, map2)

			Expect(result.Get("key1")).To(Equal("value1"))
		})

		It("handles empty second map", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"key1": "value1",
			})
			map2 := NewMap()

			result := Merge(map1, map2)

			Expect(result.Get("key1")).To(Equal("value1"))
		})

		It("handles both maps empty", func() {
			map1 := NewMap()
			map2 := NewMap()

			result := Merge(map1, map2)

			Expect(result.Count()).To(Equal(0))
		})

		It("does not modify original maps", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"key1": "value1",
			})
			map2 := NewMap(map[interface{}]interface{}{
				"key2": "value2",
			})

			result := Merge(map1, map2)
			result.Set("key3", "value3")

			Expect(map1.Has("key3")).To(BeFalse())
			Expect(map2.Has("key3")).To(BeFalse())
		})
	})

	Describe("DeepMerge", func() {
		It("merges multiple maps", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"key1": "value1",
			})
			map2 := NewMap(map[interface{}]interface{}{
				"key2": "value2",
			})
			map3 := NewMap(map[interface{}]interface{}{
				"key3": "value3",
			})

			result := DeepMerge(map1, map2, map3)

			Expect(result.Get("key1")).To(Equal("value1"))
			Expect(result.Get("key2")).To(Equal("value2"))
			Expect(result.Get("key3")).To(Equal("value3"))
		})

		It("deeply merges nested maps", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"outer": map[interface{}]interface{}{
					"inner1": "value1",
				},
			})
			map2 := NewMap(map[interface{}]interface{}{
				"outer": map[interface{}]interface{}{
					"inner2": "value2",
				},
			})

			result := DeepMerge(map1, map2)

			outer := NewMap(result.Get("outer"))
			Expect(outer.Get("inner1")).To(Equal("value1"))
			Expect(outer.Get("inner2")).To(Equal("value2"))
		})

		It("merges slices by appending", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"list": []interface{}{"item1", "item2"},
			})
			map2 := NewMap(map[interface{}]interface{}{
				"list": []interface{}{"item3", "item4"},
			})

			result := DeepMerge(map1, map2)

			list := result.Get("list").([]interface{})
			Expect(len(list)).To(Equal(4))
			Expect(list[0]).To(Equal("item1"))
			Expect(list[1]).To(Equal("item2"))
			Expect(list[2]).To(Equal("item3"))
			Expect(list[3]).To(Equal("item4"))
		})

		It("overwrites simple values from later maps", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"key": "value1",
			})
			map2 := NewMap(map[interface{}]interface{}{
				"key": "value2",
			})
			map3 := NewMap(map[interface{}]interface{}{
				"key": "value3",
			})

			result := DeepMerge(map1, map2, map3)

			Expect(result.Get("key")).To(Equal("value3"))
		})

		It("handles empty maps list", func() {
			result := DeepMerge()

			Expect(result.Count()).To(Equal(0))
		})

		It("handles single map", func() {
			map1 := NewMap(map[interface{}]interface{}{
				"key": "value",
			})

			result := DeepMerge(map1)

			Expect(result.Get("key")).To(Equal("value"))
		})
	})

	Describe("Reduce", func() {
		It("reduces collections using a reducer function", func() {
			collections := []Map{
				NewMap(map[interface{}]interface{}{"key1": "value1"}),
				NewMap(map[interface{}]interface{}{"key2": "value2"}),
				NewMap(map[interface{}]interface{}{"key3": "value3"}),
			}

			reducer := func(key, val interface{}, reducedVal Map) Map {
				reducedVal.Set(key, val)
				return reducedVal
			}

			result := Reduce(collections, NewMap(), reducer)

			Expect(result.Get("key1")).To(Equal("value1"))
			Expect(result.Get("key2")).To(Equal("value2"))
			Expect(result.Get("key3")).To(Equal("value3"))
		})

		It("passes accumulated value through reducer chain", func() {
			collections := []Map{
				NewMap(map[interface{}]interface{}{"count": 1}),
				NewMap(map[interface{}]interface{}{"count": 2}),
				NewMap(map[interface{}]interface{}{"count": 3}),
			}

			reducer := func(key, val interface{}, reducedVal Map) Map {
				if reducedVal.Has(key) {
					currentCount := reducedVal.Get(key).(int)
					reducedVal.Set(key, currentCount+val.(int))
				} else {
					reducedVal.Set(key, val)
				}
				return reducedVal
			}

			result := Reduce(collections, NewMap(), reducer)

			Expect(result.Get("count")).To(Equal(6)) // 1 + 2 + 3
		})

		It("handles empty collections", func() {
			collections := []Map{}
			reducer := func(key, val interface{}, reducedVal Map) Map {
				reducedVal.Set(key, val)
				return reducedVal
			}

			result := Reduce(collections, NewMap(), reducer)

			Expect(result.Count()).To(Equal(0))
		})

		It("works with non-empty initial value", func() {
			collections := []Map{
				NewMap(map[interface{}]interface{}{"key1": "value1"}),
			}

			initialMap := NewMap(map[interface{}]interface{}{
				"existing": "value",
			})

			reducer := func(key, val interface{}, reducedVal Map) Map {
				reducedVal.Set(key, val)
				return reducedVal
			}

			result := Reduce(collections, initialMap, reducer)

			Expect(result.Get("existing")).To(Equal("value"))
			Expect(result.Get("key1")).To(Equal("value1"))
		})

		It("can transform values during reduction", func() {
			collections := []Map{
				NewMap(map[interface{}]interface{}{"num": 5}),
				NewMap(map[interface{}]interface{}{"num": 10}),
			}

			reducer := func(key, val interface{}, reducedVal Map) Map {
				// Double the value before storing
				reducedVal.Set(key, val.(int)*2)
				return reducedVal
			}

			result := Reduce(collections, NewMap(), reducer)

			Expect(result.Get("num")).To(Equal(20)) // 10 * 2 (last value doubled)
		})
	})
})
