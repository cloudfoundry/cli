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
	source     rand.Source
	adjectives []string
	nouns      []string
}

func (wg wordGenerator) Babble() (word string) {
	idx := int(wg.source.Int63()) % len(wg.adjectives)
	word = wg.adjectives[idx] + "-"
	idx = int(wg.source.Int63()) % len(wg.nouns)
	word += wg.nouns[idx]
	return
}

func NewWordGenerator() WordGenerator {
	adjectiveBytes, _ := Asset("src/words/dict/adjectives.txt")
	nounBytes, _ := Asset("src/words/dict/nouns.txt")

	return wordGenerator{
		adjectives: strings.Split(string(adjectiveBytes), "\n"),
		nouns:      strings.Split(string(nounBytes), "\n"),
		source:     rand.NewSource(time.Now().UnixNano()),
	}
}
