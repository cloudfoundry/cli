package generic_test

import (
	. "code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("generic maps", func() {
	Describe("DeepMerge", func() {
		It("deep merges, with the last map taking precedence in conflicts", func() {
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
			Expect(mergedMap).To(Equal(expectedMap))
		})

		Context("when the second map overwrites a value in the first map with nil", func() {
			It("sets the value to nil", func() {
				map1 := NewMap(map[interface{}]interface{}{
					"nest1": map[interface{}]interface{}{},
				})

				map2 := NewMap(map[interface{}]interface{}{
					"nest1": nil,
				})

				mergedMap := DeepMerge(map1, map2)
				Expect(mergedMap.Get("nest1")).To(BeNil())
			})
		})

		Context("when the second map overwrites a nil value in the first map", func() {
			It("overwrites the nil value", func() {
				expected := map[interface{}]interface{}{}

				map1 := NewMap(map[interface{}]interface{}{
					"nest1": nil,
				})

				map2 := NewMap(map[interface{}]interface{}{
					"nest1": expected,
				})

				mergedMap := DeepMerge(map1, map2)
				Expect(mergedMap.Get("nest1")).To(BeEquivalentTo(&expected))
			})
		})
	})

	Describe("IsMappable", func() {
		It("returns true for generic.Map", func() {
			m := NewMap()
			Expect(IsMappable(m)).To(BeTrue())
		})

		It("returns true for maps", func() {
			var m map[string]interface{}
			Expect(IsMappable(m)).To(BeTrue())

			var n map[interface{}]interface{}
			Expect(IsMappable(n)).To(BeTrue())

			var o map[int]interface{}
			Expect(IsMappable(o)).To(BeTrue())
		})

		It("returns false for other things", func() {
			Expect(IsMappable(2)).To(BeFalse())
			Expect(IsMappable("2")).To(BeFalse())
			Expect(IsMappable(true)).To(BeFalse())
			Expect(IsMappable([]string{"hello"})).To(BeFalse())
			Expect(IsMappable(nil)).To(BeFalse())
		})
	})
})
