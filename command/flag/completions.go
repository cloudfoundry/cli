package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

func completions(thingsToMatch []string, prefix string) []flags.Completion {
	matches := make([]flags.Completion, 0, len(thingsToMatch))
	for _, thing := range thingsToMatch {
		if strings.HasPrefix(strings.ToLower(thing), strings.ToLower(prefix)) {
			matches = append(matches, flags.Completion{Item: thing})
		}
	}
	return matches
}
