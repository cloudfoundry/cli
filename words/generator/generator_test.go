package generator_test

import (
	"github.com/cloudfoundry/cli/words/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Word Generator", func() {
	var wordGen generator.WordGenerator

	BeforeEach(func() {
		wordGen = generator.NewWordGenerator()
	})

	Describe("NewWordGenerator", func() {
		It("creates a word generator", func() {
			Expect(wordGen).ToNot(BeNil())
		})

		It("creates different generators", func() {
			gen1 := generator.NewWordGenerator()
			gen2 := generator.NewWordGenerator()

			Expect(gen1).ToNot(BeNil())
			Expect(gen2).ToNot(BeNil())
		})
	})

	Describe("Babble", func() {
		It("generates a word", func() {
			word := wordGen.Babble()

			Expect(word).ToNot(BeEmpty())
		})

		It("generates words in adjective-noun format", func() {
			word := wordGen.Babble()

			parts := strings.Split(word, "-")
			Expect(len(parts)).To(Equal(2), "Word should be in 'adjective-noun' format")
		})

		It("generates different words on multiple calls", func() {
			words := make(map[string]bool)

			// Generate 50 words - with a good random generator,
			// we should get at least a few different words
			for i := 0; i < 50; i++ {
				word := wordGen.Babble()
				words[word] = true
			}

			// With random generation, we should have more than 1 unique word
			Expect(len(words)).To(BeNumerically(">", 1))
		})

		It("generates words without leading/trailing whitespace", func() {
			word := wordGen.Babble()

			Expect(word).To(Equal(strings.TrimSpace(word)))
		})

		It("generates words with valid characters", func() {
			word := wordGen.Babble()

			// Word should only contain letters and hyphen
			Expect(word).To(MatchRegexp(`^[a-zA-Z]+-[a-zA-Z]+$`))
		})

		It("generates multiple words successfully", func() {
			for i := 0; i < 10; i++ {
				word := wordGen.Babble()
				Expect(word).ToNot(BeEmpty())
				Expect(strings.Contains(word, "-")).To(BeTrue())
			}
		})

		It("generates words with both parts non-empty", func() {
			word := wordGen.Babble()
			parts := strings.Split(word, "-")

			Expect(parts[0]).ToNot(BeEmpty(), "Adjective part should not be empty")
			Expect(parts[1]).ToNot(BeEmpty(), "Noun part should not be empty")
		})

		It("generates consistent format across multiple calls", func() {
			for i := 0; i < 20; i++ {
				word := wordGen.Babble()
				parts := strings.Split(word, "-")

				Expect(len(parts)).To(Equal(2))
				Expect(parts[0]).ToNot(BeEmpty())
				Expect(parts[1]).ToNot(BeEmpty())
			}
		})
	})

	Describe("Multiple Generators", func() {
		It("different generators produce words", func() {
			gen1 := generator.NewWordGenerator()
			gen2 := generator.NewWordGenerator()

			word1 := gen1.Babble()
			word2 := gen2.Babble()

			Expect(word1).ToNot(BeEmpty())
			Expect(word2).ToNot(BeEmpty())
		})

		It("generators work independently", func() {
			gen1 := generator.NewWordGenerator()
			gen2 := generator.NewWordGenerator()

			// Generate from both
			word1a := gen1.Babble()
			word2a := gen2.Babble()
			word1b := gen1.Babble()
			word2b := gen2.Babble()

			// All should be valid
			Expect(word1a).ToNot(BeEmpty())
			Expect(word2a).ToNot(BeEmpty())
			Expect(word1b).ToNot(BeEmpty())
			Expect(word2b).ToNot(BeEmpty())
		})
	})

	Describe("Edge Cases", func() {
		It("handles rapid successive calls", func() {
			words := make([]string, 100)
			for i := 0; i < 100; i++ {
				words[i] = wordGen.Babble()
			}

			// All words should be valid
			for _, word := range words {
				Expect(word).ToNot(BeEmpty())
				Expect(strings.Contains(word, "-")).To(BeTrue())
			}
		})
	})
})
