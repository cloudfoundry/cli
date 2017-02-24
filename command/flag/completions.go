package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

func completions(options []string, prefix string, caseSensitive bool) []flags.Completion {
	if !caseSensitive {
		prefix = strings.ToLower(prefix)
	}

	matches := make([]flags.Completion, 0, len(options))
	for _, option := range options {
		casedOption := option
		if !caseSensitive {
			casedOption = strings.ToLower(option)
		}
		if strings.HasPrefix(casedOption, prefix) {
			matches = append(matches, flags.Completion{Item: option})
		}
	}

	return matches
}
