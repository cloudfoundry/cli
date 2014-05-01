package words

import (
	real_words "github.com/cloudfoundry/cli/words"
)

type fakeWordGenerator struct {
	fakeWord string
}

func (wg fakeWordGenerator) Babble() string {
	return wg.fakeWord
}

func NewFakeWordGenerator(fakeWord string) real_words.WordGenerator {
	return fakeWordGenerator{
		fakeWord: fakeWord,
	}
}
