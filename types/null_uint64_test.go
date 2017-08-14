package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullUint64", func() {
	var nullInt NullUint64

	BeforeEach(func() {
		nullInt = NullUint64{}
	})

	Describe("ParseFlagValue", func() {
		Context("when the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := nullInt.ParseFlagValue("")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInt).To(Equal(NullUint64{Value: 0, IsSet: false}))
			})
		})

		Context("when an invalid integer is provided", func() {
			It("returns an error", func() {
				err := nullInt.ParseFlagValue("abcdef")
				Expect(err).To(HaveOccurred())
				Expect(nullInt).To(Equal(NullUint64{Value: 0, IsSet: false}))
			})
		})

		Context("when a negative integer is provided", func() {
			It("returns an error", func() {
				err := nullInt.ParseFlagValue("-1")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a valid integer is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := nullInt.ParseFlagValue("0")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInt).To(Equal(NullUint64{Value: 0, IsSet: true}))
			})
		})
	})
})
