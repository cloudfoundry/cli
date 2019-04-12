package randomword_test

import (
	"time"

	. "code.cloudfoundry.org/cli/util/randomword"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generator", func() {
	var gen Generator

	Describe("RandomAdjective", func() {
		It("generates a random adjective each time it is called", func() {
			setOne := []string{}
			setTwo := []string{}

			for i := 0; i < 3; i++ {
				setOne = append(setOne, gen.RandomAdjective())
				// We wait for 3 nanoseconds because the seed we use to generate the
				// randomness has a unit of 1 nanosecond plus random test flakiness
				time.Sleep(3 * time.Nanosecond)
				setTwo = append(setTwo, gen.RandomAdjective())
			}
			Expect(setOne).ToNot(ConsistOf(setTwo))
		})
	})

	Describe("RandomNoun", func() {
		It("generates a random noun each time it is called", func() {
			setOne := []string{}
			setTwo := []string{}

			for i := 0; i < 3; i++ {
				setOne = append(setOne, gen.RandomNoun())
				// We wait for 3 nanoseconds because the seed we use to generate the
				// randomness has a unit of 1 nanosecond plus random test flakiness
				time.Sleep(3 * time.Nanosecond)
				setTwo = append(setTwo, gen.RandomNoun())
			}
			Expect(setOne).ToNot(ConsistOf(setTwo))
		})
	})

	Describe("Babble", func() {
		It("generates a random adjective noun pair each time it is called", func() {
			wordPair := gen.Babble()
			Expect(wordPair).To(MatchRegexp(`^\w+-\w+$`))
		})
	})
})
