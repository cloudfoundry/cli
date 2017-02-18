package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

func completions(thingsToMatch []string, prefix string) []flags.Completion {
	prefixLowered := strings.ToLower(prefix)
	matches := make([]flags.Completion, 0, len(thingsToMatch))
	for _, thing := range thingsToMatch {
		if strings.HasPrefix(strings.ToLower(thing), prefixLowered) {
			matches = append(matches, flags.Completion{Item: thing})
		}
	}
	return matches
}
