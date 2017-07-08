package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type Color struct {
	Color bool
}

func (Color) Complete(prefix string) []flags.Completion {
	return completions([]string{"true", "false"}, prefix, false)
}

func (c *Color) UnmarshalFlag(val string) error {
	switch strings.ToLower(val) {
	case "true":
		c.Color = true
	case "false":
		c.Color = false
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `COLOR must be "true" or "false"`,
		}
	}

	return nil
}
