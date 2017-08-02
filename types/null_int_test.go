package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullInt", func() {
	var nullInt NullInt

	BeforeEach(func() {
		nullInt = NullInt{}
	})

	Describe("ParseFlagValue", func() {
		Context("when the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := nullInt.ParseFlagValue("")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInt).To(Equal(NullInt{Value: 0, IsSet: false}))
			})
		})

		Context("when an invalid integer is provided", func() {
			It("returns an error", func() {
				err := nullInt.ParseFlagValue("abcdef")
				Expect(err).To(HaveOccurred())
				Expect(nullInt).To(Equal(NullInt{Value: 0, IsSet: false}))
			})
		})

		Context("when a valid integer is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := nullInt.ParseFlagValue("0")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInt).To(Equal(NullInt{Value: 0, IsSet: true}))
			})
		})
	})
})
