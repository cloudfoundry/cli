package sorting

import "unicode"

// SortAlphabeticFunc returns a `less()` comparator for sorting strings while
// respecting case.
func SortAlphabeticFunc(list []string) func(i, j int) bool {
	return func(i, j int) bool {
		iRunes := []rune(list[i])
		jRunes := []rune(list[j])

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
}
