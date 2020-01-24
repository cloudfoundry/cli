package flag

import (
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

// When setting limits allows -1 as as an unlimited value. Unlimited value is represented by nil
type IntegerLimit struct {
	types.NullInt
}


func (i *IntegerLimit) UnmarshalFlag(val string) error {
	err := i.ParseStringValue(val)
	if err != nil || i.Value < -1 {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "invalid integer limit (expected int >= -1)",
		}
	}
	return nil
}
