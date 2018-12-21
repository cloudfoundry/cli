package sorting_test

import (
	"sort"

	. "code.cloudfoundry.org/cli/util/sorting"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("SortAlphabeticFunc", func() {
	DescribeTable("sorts strings alphabetically when",
		func(input []string, expected []string) {
			sort.Slice(input, SortAlphabeticFunc(input))
			Expect(input).To(Equal(expected))
		},

		Entry("the slice is empty",
			[]string{},
			[]string{}),

		Entry("the slice contains one element",
			[]string{"a"},
			[]string{"a"}),

		Entry("the slice contains duplicates",
			[]string{"blurb", "a", "blurb"},
			[]string{"a", "blurb", "blurb"}),

		Entry("there are mixed cases and numbers",
			[]string{
				"sister",
				"Father",
				"Mother",
				"brother",
				"3-twins",
			},
			[]string{
				"3-twins",
				"brother",
				"Father",
				"Mother",
				"sister",
			}),

		Entry("capitals come before lowercase",
			[]string{
				"Stack2",
				"stack3",
				"stack1",
			},
			[]string{
				"stack1",
				"Stack2",
				"stack3",
			}),

		Entry("the strings are already sorted",
			[]string{
				"sb0",
				"sb1",
				"sb2",
			},
			[]string{
				"sb0",
				"sb1",
				"sb2",
			}),
	)
})
