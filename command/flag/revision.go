package flag

import (
	"code.cloudfoundry.org/cli/v8/types"
	flags "github.com/jessevdk/go-flags"
)

type Revision struct {
	types.NullInt
}

func (i *Revision) UnmarshalFlag(val string) error {
	err := i.ParseStringValue(val)
	if err != nil {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "invalid argument for flag '--revision' (expected int > 0)",
		}
	}
	if i.Value < 1 {
		if i.Value == 0 && i.IsSet == false {
			return nil
		}
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "invalid argument for flag '--revision' (expected int > 0)",
		}
	}
	return nil
}

func (i *Revision) IsValidValue(val string) error {
	return i.UnmarshalFlag(val)
}
