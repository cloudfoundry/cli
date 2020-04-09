package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type Color struct {
	Value string
	IsSet bool
}

func (Color) Complete(prefix string) []flags.Completion {
	return completions([]string{"true", "false"}, prefix, false)
}

func (c *Color) UnmarshalFlag(val string) error {
	switch strings.ToLower(val) {
	case "true":
		c.Value = "true"
		c.IsSet = true
	case "false":
		c.Value = "false"
		c.IsSet = true
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `COLOR must be "true" or "false"`,
		}
	}

	return nil
}
