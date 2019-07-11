package flag

import "strings"

type Tags struct {
	IsSet bool
	Value []string
}

func (t *Tags) UnmarshalFlag(value string) error {
	resultTags := []string{}

	tags := strings.Split(value, ",")
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			resultTags = append(resultTags, trimmed)
		}
	}

	t.IsSet = true
	t.Value = resultTags
	return nil
}
