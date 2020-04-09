package flag

import (
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

type Timeout struct {
	types.NullInt
}

func (t *Timeout) UnmarshalFlag(rawValue string) error {
	err := t.ParseStringValue(rawValue)
	if err != nil || t.Value < 1 {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "Timeout must be an integer greater than or equal to 1",
		}
	}
	return nil
}

func (t *Timeout) IsValidValue(val string) error {
	return t.UnmarshalFlag(val)
}
