package flag

import (
	"fmt"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type BindingName struct {
	Value string
}

func (b *BindingName) UnmarshalFlag(val string) error {
	if val == "" {
		return &flags.Error{
			Type:    flags.ErrMarshal,
			Message: "--binding-name must be at least 1 character in length",
		}
	} else if strings.HasPrefix(val, "-") {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: fmt.Sprintf("expected argument for flag --binding-name, but got option %s", val),
		}
	}

	b.Value = val
	return nil
}
