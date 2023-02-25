package types_test

import (
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("optional string slice", func() {
	It("has an unset zero value", func() {
		var s types.OptionalStringSlice
		Expect(s.IsSet).To(BeFalse())
		Expect(s.Value).To(SatisfyAll(BeEmpty(), BeNil()))
	})

	It("can be converted to a string", func() {
		Expect(types.OptionalStringSlice{}.String()).To(Equal(""))
		Expect(types.NewOptionalStringSlice().String()).To(Equal(""))
		Expect(types.NewOptionalStringSlice("foo", "bar", "baz").String()).To(Equal("foo, bar, baz"))
	})

	When("marshaling", func() {
		It("can marshal to a slice", func() {
			s := struct {
				S types.OptionalStringSlice
			}{
				S: types.NewOptionalStringSlice("foo", "bar"),
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{"S":["foo","bar"]}`))
		})

		It("can marshal to an empty slice", func() {
			s := struct {
				S types.OptionalStringSlice
			}{
				S: types.NewOptionalStringSlice(),
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{"S":[]}`))
		})

		It("can be omitted during marshaling", func() {
			var s struct {
				S types.OptionalStringSlice
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{}`))
		})
	})

	When("unmarshaling", func() {
		It("can be unmarshaled from an empty slice", func() {
			var s struct {
				S types.OptionalStringSlice
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":[]}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.IsSet).To(BeTrue())
			Expect(s.S.Value).To(SatisfyAll(BeEmpty(), BeNil()))
		})

		It("can be unmarshaled from null", func() {
			var s struct {
				S types.OptionalStringSlice
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":null}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.IsSet).To(BeTrue())
			Expect(s.S.Value).To(SatisfyAll(BeEmpty(), BeNil()))
		})
	})
})
