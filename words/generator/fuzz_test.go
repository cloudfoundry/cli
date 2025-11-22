// +build gofuzz

package generator

import (
	"strings"
	"testing"
	"unicode"
)

// FuzzBabble tests word generation for crashes and invariant violations
func FuzzBabble(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed int64) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Babble panicked: %v", r)
			}
		}()

		wordGen := NewWordGenerator()
		word := wordGen.Babble()

		// Invariant: Must not be empty
		if word == "" {
			t.Error("Babble returned empty string")
		}

		// Invariant: Must contain exactly one hyphen
		hyphens := strings.Count(word, "-")
		if hyphens != 1 {
			t.Errorf("Babble returned %q with %d hyphens, want 1", word, hyphens)
		}

		// Invariant: Must be in format adjective-noun
		parts := strings.Split(word, "-")
		if len(parts) != 2 {
			t.Errorf("Babble returned %q with %d parts, want 2", word, len(parts))
		}

		// Invariant: Both parts must be non-empty
		if len(parts) == 2 {
			if parts[0] == "" || parts[1] == "" {
				t.Errorf("Babble returned %q with empty parts", word)
			}
		}

		// Invariant: Must contain only lowercase letters and hyphen
		for _, ch := range word {
			if ch != '-' && !unicode.IsLower(ch) {
				t.Errorf("Babble returned %q containing invalid character %c", word, ch)
			}
		}

		// Invariant: Must not contain special characters
		if strings.ContainsAny(word, "!@#$%^&*()+={}[]|\\:;\"'<>,.?/~`") {
			t.Errorf("Babble returned %q containing special characters", word)
		}

		// Invariant: Must not contain numbers
		if strings.ContainsAny(word, "0123456789") {
			t.Errorf("Babble returned %q containing numbers", word)
		}

		// Invariant: Must not contain whitespace
		if strings.ContainsAny(word, " \t\n\r") {
			t.Errorf("Babble returned %q containing whitespace", word)
		}

		// Invariant: Reasonable length (between 3 and 50 characters)
		if len(word) < 3 || len(word) > 50 {
			t.Errorf("Babble returned %q with length %d, want 3-50", word, len(word))
		}
	})
}

// FuzzConcurrentBabble tests for race conditions
func FuzzConcurrentBabble(f *testing.F) {
	f.Fuzz(func(t *testing.T, numGoroutines int) {
		// Limit to reasonable range
		if numGoroutines < 1 || numGoroutines > 100 {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Concurrent Babble panicked: %v", r)
			}
		}()

		wordGen := NewWordGenerator()
		done := make(chan bool, numGoroutines)

		// Run concurrent word generation
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Goroutine panicked: %v", r)
					}
					done <- true
				}()

				for j := 0; j < 10; j++ {
					word := wordGen.Babble()
					if word == "" {
						t.Error("Concurrent Babble returned empty string")
					}
				}
			}()
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
