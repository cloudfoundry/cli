package types_test

import (
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("optional object", func() {
	It("has an unset zero value", func() {
		var s types.OptionalObject
		Expect(s.IsSet).To(BeFalse())
		Expect(s.Value).To(SatisfyAll(BeEmpty(), BeNil()))
	})

	It("has a set empty value", func() {
		s := types.NewOptionalObject(map[string]interface{}{})
		Expect(s.IsSet).To(BeTrue())
		Expect(s.Value).To(BeEmpty())
		Expect(s.Value).NotTo(BeNil())

		t := types.NewOptionalObject(nil)
		Expect(t.IsSet).To(BeTrue())
		Expect(t.Value).To(BeEmpty())
		Expect(t.Value).NotTo(BeNil())

		Expect(s).To(Equal(t))
	})

	When("marshaling", func() {
		It("can marshal to an object", func() {
			s := struct {
				S types.OptionalObject
			}{
				S: types.NewOptionalObject(map[string]interface{}{"foo": "bar"}),
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{"S":{"foo":"bar"}}`))
		})

		It("can marshal to an empty object", func() {
			s := struct {
				S types.OptionalObject
			}{
				S: types.NewOptionalObject(map[string]interface{}{}),
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{"S":{}}`))
		})

		It("can be omitted during marshaling", func() {
			var s struct {
				S types.OptionalObject
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{}`))
		})
	})

	When("unmarshaling", func() {
		It("can be unmarshaled from an empty object", func() {
			var s struct {
				S types.OptionalObject
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":{}}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.IsSet).To(BeTrue())
			Expect(s.S.Value).To(BeEmpty())
			Expect(s.S.Value).NotTo(BeNil())
		})

		It("can be unmarshaled from null", func() {
			var s struct {
				S types.OptionalObject
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":null}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.IsSet).To(BeTrue())
			Expect(s.S.Value).To(BeEmpty())
			Expect(s.S.Value).NotTo(BeNil())
		})
	})
})
