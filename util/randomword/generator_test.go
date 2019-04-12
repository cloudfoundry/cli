package randomword_test

import (
	. "code.cloudfoundry.org/cli/util/randomword"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generator", func() {
	var gen Generator

	Describe("RandomAdjective", func() {
		It("generates a random adjective each time it is called", func() {
			Eventually(gen.RandomAdjective).ShouldNot(Equal(gen.RandomAdjective()))
		})
	})

	Describe("RandomNoun", func() {
		It("generates a random noun each time it is called", func() {
			Eventually(gen.RandomNoun).ShouldNot(Equal(gen.RandomNoun()))
		})
	})

	Describe("Babble", func() {
		It("generates a random adjective noun pair each time it is called", func() {
			wordPair := gen.Babble()
			Expect(wordPair).To(MatchRegexp(`^\w+-\w+$`))
		})
	})
})
