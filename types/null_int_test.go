package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullInt", func() {
	var nullInt NullInt

	BeforeEach(func() {
		nullInt = NullInt{}
	})

	Describe("ParseIntValue", func() {
		Context("when nil is provided", func() {
			It("sets IsSet to false", func() {
				nullInt.ParseIntValue(nil)
				Expect(nullInt).To(Equal(NullInt{Value: 0, IsSet: false}))
			})
		})

		Context("when non-nil pointer is provided", func() {
			It("sets IsSet to true and Value to provided value", func() {
				n := 5
				nullInt.ParseIntValue(&n)
				Expect(nullInt).To(Equal(NullInt{Value: 5, IsSet: true}))
			})
		})
	})

	Describe("ParseStringValue", func() {
		Context("when the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := nullInt.ParseStringValue("")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInt).To(Equal(NullInt{Value: 0, IsSet: false}))
			})
		})

		Context("when an invalid integer is provided", func() {
			It("returns an error", func() {
				err := nullInt.ParseStringValue("abcdef")
				Expect(err).To(HaveOccurred())
				Expect(nullInt).To(Equal(NullInt{Value: 0, IsSet: false}))
			})
		})

		Context("when a valid integer is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := nullInt.ParseStringValue("0")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInt).To(Equal(NullInt{Value: 0, IsSet: true}))
			})
		})
	})

	Describe("UnmarshalJSON", func() {
		Context("when integer value is provided", func() {
			It("parses JSON number correctly", func() {
				err := nullInt.UnmarshalJSON([]byte("42"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInt).To(Equal(NullInt{Value: 42, IsSet: true}))
			})
		})

		Context("when empty json is provided", func() {
			It("returns an unset NullInt", func() {
				err := nullInt.UnmarshalJSON([]byte(`""`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInt).To(Equal(NullInt{Value: 0, IsSet: false}))
			})
		})
	})

	DescribeTable("MarshalJSON",
		func(nullInt NullInt, expectedBytes []byte) {
			bytes, err := nullInt.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(Equal(expectedBytes))
		},
		Entry("negative number", NullInt{IsSet: true, Value: -1}, []byte("-1")),
		Entry("positive number", NullInt{IsSet: true, Value: 1}, []byte("1")),
		Entry("0", NullInt{IsSet: true, Value: 0}, []byte("0")),
		Entry("no value", NullInt{IsSet: false}, []byte("null")),
	)
})
