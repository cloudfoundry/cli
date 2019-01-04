package flag

import (
	"strings"
)

type Tags []string

func (t *Tags) UnmarshalFlag(value string) error {
	resultTags := []string{}

	tags := strings.Split(value, ",")
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			resultTags = append(resultTags, trimmed)
		}
	}

	*t = Tags(resultTags)
	return nil
}
