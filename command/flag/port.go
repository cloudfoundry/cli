package flag

import (
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

type Port struct {
	types.NullInt
}

func (i *Port) UnmarshalFlag(val string) error {
	err := i.ParseStringValue(val)
	if err != nil || i.Value < 0 {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "invalid argument for flag '--port' (expected int > 0)",
		}
	}
	return nil
}

// IsValidValue returns an error if the input value is not an integer.
func (i *Port) IsValidValue(val string) error {
	return i.UnmarshalFlag(val)
}
