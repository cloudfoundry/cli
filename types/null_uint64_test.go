package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullUint64", func() {
	var nullUint64 NullUint64

	BeforeEach(func() {
		nullUint64 = NullUint64{}
	})

	Describe("ParseStringValue", func() {
		Context("when the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := nullUint64.ParseStringValue("")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullUint64).To(Equal(NullUint64{Value: 0, IsSet: false}))
			})
		})

		Context("when an invalid integer is provided", func() {
			It("returns an error", func() {
				err := nullUint64.ParseStringValue("abcdef")
				Expect(err).To(HaveOccurred())
				Expect(nullUint64).To(Equal(NullUint64{Value: 0, IsSet: false}))
			})
		})

		Context("when a negative integer is provided", func() {
			It("returns an error", func() {
				err := nullUint64.ParseStringValue("-1")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a valid integer is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := nullUint64.ParseStringValue("0")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullUint64).To(Equal(NullUint64{Value: 0, IsSet: true}))
			})
		})
	})

	Describe("UnmarshalJSON", func() {
		Context("when integer value is provided", func() {
			It("parses JSON number correctly", func() {
				err := nullUint64.UnmarshalJSON([]byte("42"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullUint64).To(Equal(NullUint64{Value: 42, IsSet: true}))
			})
		})

		Context("when empty json is provided", func() {
			It("returns an unset NullUint64", func() {
				err := nullUint64.UnmarshalJSON([]byte(`""`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullUint64).To(Equal(NullUint64{Value: 0, IsSet: false}))
			})
		})
	})
})
