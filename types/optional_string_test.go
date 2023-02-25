package types_test

import (
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("optional string", func() {
	It("has an unset zero value", func() {
		var s types.OptionalString
		Expect(s.IsSet).To(BeFalse())
		Expect(s.Value).To(BeEmpty())
	})

	It("can be converted to a string", func() {
		Expect(types.OptionalString{}.String()).To(Equal(""))
		Expect(types.NewOptionalString("").String()).To(Equal(""))
		Expect(types.NewOptionalString("foo").String()).To(Equal("foo"))
	})

	When("marshaling", func() {
		It("can marshal to a string", func() {
			s := struct {
				S types.OptionalString
			}{
				S: types.NewOptionalString("foo"),
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{"S":"foo"}`))
		})

		It("can marshal to an empty string", func() {
			s := struct {
				S types.OptionalString
			}{
				S: types.NewOptionalString(""),
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{"S":""}`))
		})

		It("can be omitted during marshaling", func() {
			var s struct {
				S types.OptionalString
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{}`))
		})
	})

	When("unmarshaling", func() {
		It("can be unmarshaled from an empty string", func() {
			var s struct {
				S types.OptionalString
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":""}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.IsSet).To(BeTrue())
			Expect(s.S.Value).To(BeEmpty())
		})

		It("can be unmarshaled from null", func() {
			var s struct {
				S types.OptionalString
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":null}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.IsSet).To(BeTrue())
			Expect(s.S.Value).To(BeEmpty())
		})
	})
})
