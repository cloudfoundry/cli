package cmd

import (
	"strings"
)

type TrimmedSpaceArgs []string

func (as TrimmedSpaceArgs) AsStrings() []string {
	result := []string{}
	for _, a := range as {
		result = append(result, strings.TrimSpace(a))
	}
	return result
}
