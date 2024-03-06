package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullByteSizeInMb", func() {
	var nullByteSize NullByteSizeInMb

	BeforeEach(func() {
		nullByteSize = NullByteSizeInMb{
			IsSet: true,
			Value: 0xBAD,
		}
	})

	Describe("String", func() {
		When("a NullByteSize value is not set", func() {
			It("returns an empty string", func() {
				nullByteSize.IsSet = false
				returnedString := nullByteSize.String()
				Expect(returnedString).To(Equal(""))
			})
		})

		When("a NullByteSize value is set", func() {
			When("NullByteSize value is in megabytes", func() {
				It("returns a formatted byte size", func() {
					nullByteSize.IsSet = true
					nullByteSize.Value = 1024
					returnedString := nullByteSize.String()
					Expect(returnedString).To(Equal("1G"))
				})
			})
		})
	})

	Describe("ParseStringValue", func() {
		When("the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := nullByteSize.ParseStringValue("")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullByteSize).To(Equal(NullByteSizeInMb{Value: 0, IsSet: false}))
			})
		})

		When("an invalid byte size is provided", func() {
			It("returns an error", func() {
				err := nullByteSize.ParseStringValue("abcdef")
				Expect(err).To(HaveOccurred())
			})
		})

		When("a unit size is not provided", func() {
			It("returns an error", func() {
				err := nullByteSize.ParseStringValue("1")
				Expect(err).To(HaveOccurred())
			})
		})

		When("a valid byte size is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := nullByteSize.ParseStringValue("1G")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullByteSize).To(Equal(NullByteSizeInMb{Value: 1024, IsSet: true}))
			})
		})
	})

	Describe("ParseUint64Value", func() {
		When("nil is provided", func() {
			It("sets IsSet to false", func() {
				nullByteSize.ParseUint64Value(nil)
				Expect(nullByteSize).To(Equal(NullByteSizeInMb{Value: 0, IsSet: false}))
			})
		})

		When("non-nil pointer is provided", func() {
			It("sets IsSet to true and Value to provided value", func() {
				n := uint64(5)
				nullByteSize.ParseUint64Value(&n)
				Expect(nullByteSize).To(Equal(NullByteSizeInMb{Value: 5, IsSet: true}))
			})
		})
	})

	Describe("UnmarshalJSON", func() {
		When("integer value is provided", func() {
			It("parses JSON number correctly", func() {
				err := nullByteSize.UnmarshalJSON([]byte("42"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullByteSize).To(Equal(NullByteSizeInMb{Value: 42, IsSet: true}))
			})
		})

		When("a null value is provided", func() {
			It("returns an unset NullByteSizeInMb", func() {
				err := nullByteSize.UnmarshalJSON([]byte("null"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullByteSize).To(Equal(NullByteSizeInMb{Value: 0, IsSet: false}))
			})
		})

		When("empty string is provided", func() {
			It("returns an unset NullByteSizeInMb", func() {
				err := nullByteSize.UnmarshalJSON([]byte(""))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullByteSize).To(Equal(NullByteSizeInMb{Value: 0, IsSet: false}))
			})
		})
	})
})
