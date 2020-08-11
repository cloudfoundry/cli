package types_test

import (
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("trimmed string", func() {
	It("has an unset zero value", func() {
		var s types.TrimmedString
		Expect(s.Value).To(BeEmpty())
	})

	It("can be converted to a string", func() {
		Expect(types.TrimmedString{}.String()).To(Equal(""))
		Expect(types.NewTrimmedString("").String()).To(Equal(""))
		Expect(types.NewTrimmedString("  foo  ").String()).To(Equal("foo"))
	})

	When("marshaling", func() {
		It("can marshal to a string", func() {
			s := struct {
				S types.TrimmedString
			}{
				S: types.NewTrimmedString("   foo   "),
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{"S":"foo"}`))
		})

		It("can marshal to an empty string", func() {
			s := struct {
				S types.TrimmedString
			}{
				S: types.NewTrimmedString(""),
			}
			Expect(jsonry.Marshal(s)).To(MatchJSON(`{"S":""}`))
		})
	})

	When("unmarshalling", func() {
		It("can be unmarshalled from an empty string", func() {
			var s struct {
				S types.TrimmedString
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":""}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.Value).To(BeEmpty())
		})

		It("can be unmarshalled from null", func() {
			var s struct {
				S types.TrimmedString
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":null}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.Value).To(BeEmpty())
		})

		It("removes trailing spaces when unmarshalling", func() {
			var s struct {
				S types.TrimmedString
			}
			Expect(jsonry.Unmarshal([]byte(`{"S":"   foo   "}`), &s)).NotTo(HaveOccurred())
			Expect(s.S.Value).To(Equal("foo"))
		})
	})
})
