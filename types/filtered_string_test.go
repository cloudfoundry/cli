package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("FilteredString", func() {
	DescribeTable("ParseValue",
		func(input string, expected FilteredString) {
			nullString := FilteredString{}
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
})
