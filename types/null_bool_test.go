package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullBool", func() {
	var nullBool NullBool

	BeforeEach(func() {
		nullBool = NullBool{}
	})

	Describe("ParseBoolValue", func() {
		When("nil is provided", func() {
			It("sets IsSet to false", func() {
				nullBool.ParseBoolValue(nil)
				Expect(nullBool).To(Equal(NullBool{Value: false, IsSet: false}))
			})
		})

		When("non-nil pointer is provided", func() {
			It("sets IsSet to true and Value to provided value", func() {
				n := true
				nullBool.ParseBoolValue(&n)
				Expect(nullBool).To(Equal(NullBool{Value: true, IsSet: true}))
			})
		})
	})

	Describe("ParseStringValue", func() {
		When("the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := nullBool.ParseStringValue("")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullBool).To(Equal(NullBool{Value: false, IsSet: false}))
			})
		})

		When("an invalid integer is provided", func() {
			It("returns an error", func() {
				err := nullBool.ParseStringValue("abcdef")
				Expect(err).To(HaveOccurred())
				Expect(nullBool).To(Equal(NullBool{Value: false, IsSet: false}))
			})
		})

		When("a valid integer is provided", func() {
			It("stores the bool and sets IsSet to true", func() {
				err := nullBool.ParseStringValue("true")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullBool).To(Equal(NullBool{Value: true, IsSet: true}))
			})
		})
	})

	Describe("UnmarshalJSON", func() {
		When("integer value is provided", func() {
			It("parses JSON number correctly", func() {
				err := nullBool.UnmarshalJSON([]byte("true"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullBool).To(Equal(NullBool{Value: true, IsSet: true}))
			})
		})

		When("empty json is provided", func() {
			It("returns an unset NullBool", func() {
				err := nullBool.UnmarshalJSON([]byte("null"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullBool).To(Equal(NullBool{Value: false, IsSet: false}))
			})
		})
	})

	DescribeTable("MarshalJSON",
		func(nullBool NullBool, expectedBytes []byte) {
			bytes, err := nullBool.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(Equal(expectedBytes))
		},
		Entry("true when set", NullBool{IsSet: true, Value: true}, []byte("true")),
		Entry("false when set", NullBool{IsSet: true, Value: false}, []byte("false")),
		Entry("no value", NullBool{IsSet: false}, []byte("null")),
	)
})
