package flag

import (
	"strings"

	"code.cloudfoundry.org/cli/v8/types"
)

type Tags types.OptionalStringSlice

func (t *Tags) UnmarshalFlag(value string) error {
	tags := strings.Split(value, ",")
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			t.Value = append(t.Value, trimmed)
		}
	}

	t.IsSet = true
	return nil
}
