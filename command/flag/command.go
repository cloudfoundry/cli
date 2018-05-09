package flag

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

type Command struct {
	types.FilteredString
}

func (b *Command) UnmarshalFlag(val string) error {
	if strings.HasPrefix(val, "-") {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: fmt.Sprintf("expected argument for flag -c, but got option %s", val),
		}
	}
	b.ParseValue(val)
	return nil
}
