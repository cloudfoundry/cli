package matchers_test

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContainElementTimes", func() {
	It("asserts correctly", func() {
		array := []string{"one", "two", "three", "two", "five"}

		Expect(array).To(ContainElementTimes("two", 2))
	})

	It("handles missing elements", func() {
		array := []string{"one", "two", "three", "four", "five"}

		Expect(array).To(ContainElementTimes("something-else", 0))
		Expect(array).NotTo(ContainElementTimes("something-else", 1))
	})

	It("handles non-string types", func() {
		type FancyString string

		array := []FancyString{
			FancyString("one"),
			FancyString("two"),
			FancyString("three"),
			FancyString("four"),
			FancyString("five"),
		}

		Expect(array).To(ContainElementTimes(FancyString("four"), 1))
	})

	It("errors when given bad data", func() {
		matcher := ContainElementTimesMatcher{Element: 1}
		_, err := matcher.Match(2)

		Expect(err).To(MatchError("expected an array"))
	})
})
