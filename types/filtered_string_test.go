package types_test

import (
	. "code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FilteredString", func() {
	var nullString FilteredString

	DescribeTable("ParseValue",
		func(input string, expected FilteredString) {
			nullString.ParseValue(input)
			Expect(nullString).To(Equal(expected))
		},

		Entry("empty string", "", FilteredString{}),
		Entry("default", "default", FilteredString{IsSet: true}),
		Entry("null", "null", FilteredString{IsSet: true}),
		Entry("some other string", "literally-anything-else", FilteredString{
			IsSet: true,
			Value: "literally-anything-else",
		}),
	)

	DescribeTable("IsDefault",
		func(input string, expected bool) {
			nullString.ParseValue(input)
			Expect(nullString.IsDefault()).To(Equal(expected))
		},

		Entry("empty string returns false", "", false),
		Entry("default returns true", "default", true),
		Entry("null returns true", "null", true),
		Entry("some other string returns false", "literally-anything-else", false),
	)

	Describe("UnmarshalJSON", func() {
		When("a string value is provided", func() {
			It("parses a out a valid FilteredString", func() {
				err := nullString.UnmarshalJSON([]byte(`"some-string"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullString).To(Equal(FilteredString{Value: "some-string", IsSet: true}))
			})
		})

		When("an empty string value is provided", func() {
			It("parses a out a valid FilteredString", func() {
				err := nullString.UnmarshalJSON([]byte(`""`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullString).To(Equal(FilteredString{Value: "", IsSet: true}))
			})
		})

		When("an empty JSON is provided", func() {
			It("parses a out a valid FilteredString", func() {
				err := nullString.UnmarshalJSON([]byte("null"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullString).To(Equal(FilteredString{Value: "", IsSet: false}))
			})
		})
	})

	Describe("MarshalJSON", func() {
		When("a FilteredString is set to some string", func() {
			It("returns a string", func() {
				nullString = FilteredString{Value: "some-string", IsSet: true}
				marshalled, err := nullString.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte(`"some-string"`)))
			})
		})

		When("a FilteredString is set to an empty string", func() {
			It("returns null", func() {
				nullString = FilteredString{Value: "", IsSet: true}
				marshalled, err := nullString.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte("null")))
			})
		})

		When("a FilteredString is not set", func() {
			It("returns null", func() {
				nullString = FilteredString{Value: "", IsSet: false}
				marshalled, err := nullString.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte("null")))
			})
		})
	})
})
