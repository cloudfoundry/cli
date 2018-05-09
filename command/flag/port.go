package flag

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

type Port struct {
	types.NullInt
}

func (i *Port) UnmarshalFlag(val string) error {
	if strings.HasPrefix(val, "-") {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: fmt.Sprintf("expected argument for flag --port, but got option %s", val),
		}
	}
	err := i.ParseStringValue(val)
	if err != nil {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "invalid argument for flag '--port' (expected int > 0)",
		}
	}
	return nil
}
