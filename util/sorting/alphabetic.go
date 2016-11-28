package sorting

import "unicode"

// Alphabetic is an array of strings that can be alphabetically sorted
type Alphabetic []string

func (s Alphabetic) Len() int      { return len(s) }
func (s Alphabetic) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s Alphabetic) Less(i, j int) bool {
	return SortAlphabetic(s[i], s[j])
}

// SortAlphabetic will return true if string a comes after string b
func SortAlphabetic(a string, b string) bool {
	iRunes := []rune(a)
	jRunes := []rune(b)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	return false
}
