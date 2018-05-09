package flag

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

type Buildpack struct {
	types.FilteredString
}

func (b *Buildpack) UnmarshalFlag(val string) error {
	if strings.HasPrefix(val, "-") {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: fmt.Sprintf("expected argument for flag --buildpack, but got option %s", val),
		}
	}
	b.ParseValue(val)
	return nil
}
