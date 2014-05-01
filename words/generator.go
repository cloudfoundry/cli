package words

import (
	"math/rand"
	"strings"
	"time"
)

type WordGenerator interface {
	Babble() string
}

type wordGenerator struct {
	numberGenerator *rand.Rand
	adjectives      []string
	nouns           []string
}

func (wg wordGenerator) Babble() (word string) {
	idx := int(wg.numberGenerator.Int()) % len(wg.adjectives)
	word = wg.adjectives[idx] + "-"
	idx = int(wg.numberGenerator.Int()) % len(wg.nouns)
	word += wg.nouns[idx]
	return
}

func NewWordGenerator() WordGenerator {
	adjectiveBytes, _ := Asset("src/words/dict/adjectives.txt")
	nounBytes, _ := Asset("src/words/dict/nouns.txt")
	source := rand.NewSource(time.Now().UnixNano())

	return wordGenerator{
		adjectives:      strings.Split(string(adjectiveBytes), "\n"),
		nouns:           strings.Split(string(nounBytes), "\n"),
		numberGenerator: rand.New(source),
	}
}
