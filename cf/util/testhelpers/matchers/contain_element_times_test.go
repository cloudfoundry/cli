package matchers_test

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ContainElementsInOrder", func() {
	It("asserts correctly", func() {
		array := []string{"one", "two", "three", "four", "five"}

		Expect(array).To(ContainElementsInOrder("two", "four"))
	})

	It("handles out-of-order elements", func() {
		array := []string{"one", "two", "three", "four", "five"}

		Expect(array).NotTo(ContainElementsInOrder("four", "two"))
	})

	It("handles missing elements", func() {
		array := []string{"one", "two", "three", "four", "five"}

		Expect(array).NotTo(ContainElementsInOrder("something-else"))
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

		Expect(array).To(ContainElementsInOrder(FancyString("one"), FancyString("four")))
	})

	It("errors when given bad data", func() {
		array := []string{"1", "2"}
		matcher := ContainElementsInOrderMatcher{Elements: array}
		_, err := matcher.Match(2)

		Expect(err).To(MatchError("expected an array"))
	})
})
