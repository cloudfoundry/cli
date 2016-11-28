package sorting_test

import (
	"sort"

	. "code.cloudfoundry.org/cli/util/sorting"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Alphabetic", func() {
	It("sorts an empty slice", func() {
		sample := []string{}
		sort.Sort(Alphabetic(sample))
		Expect(sample).To(Equal([]string{}))
	})

	It("sorts a slice of size 1", func() {
		sample := []string{"a"}
		sort.Sort(Alphabetic(sample))
		Expect(sample).To(Equal([]string{"a"}))
	})

	It("sorts a duplicates", func() {
		sample := []string{"blurb", "blurb"}
		sort.Sort(Alphabetic(sample))
		Expect(sample).To(Equal([]string{"blurb", "blurb"}))
	})

	It("sorts strings alphabetically regardless of case", func() {
		sample := []string{
			"sister",
			"Father",
			"Mother",
			"brother",
			"3-twins",
		}
		expected := []string{
			"3-twins",
			"brother",
			"Father",
			"Mother",
			"sister",
		}
		sort.Sort(Alphabetic(sample))
		Expect(sample).To(Equal(expected))
	})
})
