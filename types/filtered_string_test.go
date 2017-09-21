package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

	Describe("UnmarshalJSON", func() {
		Context("when a string value is provided", func() {
			It("parses a out a valid FilteredString", func() {
				err := nullString.UnmarshalJSON([]byte(`"some-string"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullString).To(Equal(FilteredString{Value: "some-string", IsSet: true}))
			})
		})

		Context("when an empty string value is provided", func() {
			It("parses a out a valid FilteredString", func() {
				err := nullString.UnmarshalJSON([]byte(`""`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullString).To(Equal(FilteredString{Value: "", IsSet: true}))
			})
		})

		Context("when an empty JSON is provided", func() {
			It("parses a out a valid FilteredString", func() {
				err := nullString.UnmarshalJSON([]byte("null"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullString).To(Equal(FilteredString{Value: "", IsSet: false}))
			})
		})
	})

	Describe("MarshalJSON", func() {
		Context("when a FilteredString is set to some string", func() {
			It("returns a string", func() {
				nullString = FilteredString{Value: "some-string", IsSet: true}
				marshalled, err := nullString.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte(`"some-string"`)))
			})
		})

		Context("when a FilteredString is set to an empty string", func() {
			It("returns an empty string", func() {
				nullString = FilteredString{Value: "", IsSet: true}
				marshalled, err := nullString.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte(`""`)))
			})
		})

		Context("when a FilteredString is not set", func() {
			It("returns null", func() {
				nullString = FilteredString{Value: "", IsSet: false}
				marshalled, err := nullString.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte("null")))
			})
		})
	})
})
