package uihelpers_test

import (
	. "code.cloudfoundry.org/cli/cf/uihelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("tags parser", func() {

	It("parses an empty string", func() {
		rawTag := ""

		Expect(ParseTags(rawTag)).To(Equal([]string{}))
	})

	It("parses a single tag string", func() {
		rawTag := "a, b, c, d"

		Expect(ParseTags(rawTag)).To(Equal([]string{"a", "b", "c", "d"}))
	})

	Context("and the formatting isn't a perfect comma-delimited list", func() {

		It("parses a single tag string", func() {
			rawTag := "a,b, c,d"

			Expect(ParseTags(rawTag)).To(Equal([]string{"a", "b", "c", "d"}))
		})

		It("parses a single tag string", func() {
			rawTag := " a, b, c, d "

			Expect(ParseTags(rawTag)).To(Equal([]string{"a", "b", "c", "d"}))
		})

		It("parses a single tag string", func() {
			rawTag := "a"

			Expect(ParseTags(rawTag)).To(Equal([]string{"a"}))
		})

		It("parses a single tag string", func() {
			rawTag := ",,,,,a,,,,,b"

			Expect(ParseTags(rawTag)).To(Equal([]string{"a", "b"}))
		})

		It("parses a single tag string", func() {
			rawTag := "a, , , b"

			Expect(ParseTags(rawTag)).To(Equal([]string{"a", "b"}))
		})
	})
})
