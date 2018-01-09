package randomword_test

import (
	"time"

	. "code.cloudfoundry.org/cli/util/randomword"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generator", func() {
	var gen Generator

	BeforeEach(func() {
		gen = Generator{}
	})

	Describe("RandomAdjective", func() {
		It("generates a random adjective each time it is called", func() {
			adj := gen.RandomAdjective()
			// We wait for 3 millisecond because the seed we use to generate the
			// randomness has a unit of 1 nanosecond plus random test flakiness
			time.Sleep(3)
			Expect(adj).ToNot(Equal(gen.RandomAdjective()))
		})
	})

	Describe("RandomNoun", func() {
		It("generates a random noun each time it is called", func() {
			noun := gen.RandomNoun()
			// We wait for 3 millisecond because the seed we use to generate the
			// randomness has a unit of 1 nanosecond plus random test flakiness
			time.Sleep(10)
			Expect(noun).ToNot(Equal(gen.RandomNoun()))
		})
	})

	Describe("Babble", func() {
		It("generates a random adjective noun pair each time it is called", func() {
			wordPair := gen.Babble()
			Expect(wordPair).To(MatchRegexp("^\\w+-\\w+$"))
		})
	})
})
