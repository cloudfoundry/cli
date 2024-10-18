package types_test

import (
	. "code.cloudfoundry.org/cli/v8/types"
	"github.com/jessevdk/go-flags"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullFloat64", func() {
	var nullFloat64 NullFloat64

	BeforeEach(func() {
		nullFloat64 = NullFloat64{
			IsSet: true,
			Value: 0xBAD,
		}
	})

	Describe("IsValidValue", func() {
		var (
			input      string
			executeErr error
		)

		JustBeforeEach(func() {
			executeErr = nullFloat64.IsValidValue(input)
		})

		When("the value is a positive float", func() {
			BeforeEach(func() {
				input = "1.01"
			})

			It("does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})

		When("the value is a negative float", func() {
			BeforeEach(func() {
				input = "-21.94"
			})

			It("does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})

		When("the value is a non float", func() {
			BeforeEach(func() {
				input = "not-a-integer"
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("invalid float64 value `not-a-integer`"))
			})
		})
	})

	Describe("ParseFloat64Value", func() {
		When("nil is provided", func() {
			It("sets IsSet to false", func() {
				nullFloat64.ParseFloat64Value(nil)
				Expect(nullFloat64).To(Equal(NullFloat64{Value: 0, IsSet: false}))
			})
		})

		When("non-nil pointer is provided", func() {
			It("sets IsSet to true and Value to provided value", func() {
				n := 5.04
				nullFloat64.ParseFloat64Value(&n)
				Expect(nullFloat64).To(Equal(NullFloat64{Value: 5.04, IsSet: true}))
			})
		})
	})

	Describe("ParseStringValue", func() {
		When("the empty string is provided", func() {
			It("sets IsSet to false", func() {
				err := nullFloat64.ParseStringValue("")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullFloat64).To(Equal(NullFloat64{Value: 0, IsSet: false}))
			})
		})

		When("an invalid float64 is provided", func() {
			It("returns an error", func() {
				err := nullFloat64.ParseStringValue("abcdef")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrMarshal,
					Message: "invalid float64 value `abcdef`",
				}))
				Expect(nullFloat64).To(Equal(NullFloat64{Value: 0, IsSet: false}))
			})
		})

		When("a valid float64 is provided", func() {
			It("stores the integer and sets IsSet to true", func() {
				err := nullFloat64.ParseStringValue("0")
				Expect(err).ToNot(HaveOccurred())
				Expect(nullFloat64).To(Equal(NullFloat64{Value: 0, IsSet: true}))
			})
		})
	})

	Describe("UnmarshalJSON", func() {
		When("float64 value is provided", func() {
			It("parses JSON number correctly", func() {
				err := nullFloat64.UnmarshalJSON([]byte("42.333333"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullFloat64).To(Equal(NullFloat64{Value: 42.333333, IsSet: true}))
			})
		})

		When("a null value is provided", func() {
			It("returns an unset NullFloat64", func() {
				err := nullFloat64.UnmarshalJSON([]byte("null"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullFloat64).To(Equal(NullFloat64{Value: 0, IsSet: false}))
			})
		})

		When("an empty string is provided", func() {
			It("returns an unset NullFloat64", func() {
				err := nullFloat64.UnmarshalJSON([]byte(""))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullFloat64).To(Equal(NullFloat64{Value: 0, IsSet: false}))
			})
		})
	})

	DescribeTable("MarshalJSON",
		func(nullFloat64 NullFloat64, expectedBytes []byte) {
			bytes, err := nullFloat64.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(Equal(expectedBytes))
		},
		Entry("negative number", NullFloat64{IsSet: true, Value: -1.5}, []byte("-1.5")),
		Entry("positive number", NullFloat64{IsSet: true, Value: 1.8}, []byte("1.8")),
		Entry("0", NullFloat64{IsSet: true, Value: 0.0}, []byte("0")),
		Entry("no value", NullFloat64{IsSet: false}, []byte("null")),
	)
})
