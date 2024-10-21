package flag

import (
	"code.cloudfoundry.org/cli/v7/types"
	flags "github.com/jessevdk/go-flags"
)

type Instances struct {
	types.NullInt
}

func (i *Instances) UnmarshalFlag(val string) error {
	err := i.ParseStringValue(val)
	if err != nil || i.Value < 0 {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "invalid argument for flag '-i' (expected int > 0)",
		}
	}
	return nil
}

func (i *Instances) IsValidValue(val string) error {
	return i.UnmarshalFlag(val)
}
