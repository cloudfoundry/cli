package generator_test

import (
	"strings"
	"testing"
	"testing/quick"

	"github.com/cloudfoundry/cli/words/generator"
)

// TestBabbleAlwaysGeneratesValidFormat verifies Babble always produces valid format
func TestBabbleAlwaysGeneratesValidFormat(t *testing.T) {
	f := func(seed int) bool {
		wordGen := generator.NewWordGenerator()

		// Generate a word
		word := wordGen.Babble()

		// Should contain exactly one hyphen
		parts := strings.Split(word, "-")
		if len(parts) != 2 {
			return false
		}

		// Both parts should be non-empty
		if len(parts[0]) == 0 || len(parts[1]) == 0 {
			return false
		}

		// Both parts should be lowercase
		adjective := parts[0]
		noun := parts[1]

		return adjective == strings.ToLower(adjective) &&
			noun == strings.ToLower(noun)
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestBabbleProducesUniqueWords verifies multiple calls produce different words eventually
func TestBabbleProducesUniqueWords(t *testing.T) {
	f := func(seed int) bool {
		wordGen := generator.NewWordGenerator()

		// Generate 10 words
		words := make(map[string]bool)
		for i := 0; i < 10; i++ {
			word := wordGen.Babble()
			words[word] = true
		}

		// With random generation, we should get at least 2 unique words out of 10
		// (being conservative to avoid flaky tests)
		return len(words) >= 2
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 20}); err != nil {
		t.Error(err)
	}
}

// TestBabbleNeverPanics verifies Babble never panics
func TestBabbleNeverPanics(t *testing.T) {
	f := func(iterations uint8) bool {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Babble panicked: %v", r)
			}
		}()

		wordGen := generator.NewWordGenerator()

		// Call Babble multiple times
		for i := uint8(0); i < iterations; i++ {
			_ = wordGen.Babble()
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error(err)
	}
}

// TestBabbleConsistentLength verifies generated words have reasonable length
func TestBabbleConsistentLength(t *testing.T) {
	f := func(seed int) bool {
		wordGen := generator.NewWordGenerator()
		word := wordGen.Babble()

		// Word should be at least 3 characters (a-b format minimum)
		// and at most 50 characters (reasonable upper bound)
		length := len(word)
		return length >= 3 && length <= 50
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestBabbleOnlyContainsValidCharacters verifies only lowercase letters and hyphen
func TestBabbleOnlyContainsValidCharacters(t *testing.T) {
	f := func(seed int) bool {
		wordGen := generator.NewWordGenerator()
		word := wordGen.Babble()

		for _, ch := range word {
			// Only lowercase letters and hyphen are valid
			isLowercase := ch >= 'a' && ch <= 'z'
			isHyphen := ch == '-'

			if !isLowercase && !isHyphen {
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error(err)
	}
}

// TestNewWordGeneratorNeverPanics verifies constructor never panics
func TestNewWordGeneratorNeverPanics(t *testing.T) {
	f := func(iterations uint8) bool {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewWordGenerator panicked: %v", r)
			}
		}()

		for i := uint8(0); i < iterations; i++ {
			_ = generator.NewWordGenerator()
		}

		return true
	}

	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error(err)
	}
}
