package flag

import (
	"strconv"

	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

// When setting limits allows -1 as as an unlimited value. Unlimited value is represented by nil
type IntegerLimit types.NullInt

func (i *IntegerLimit) UnmarshalFlag(val string) error {
	if val == "" {
		return nil
	}

	intVal, err := strconv.Atoi(val)

	if err != nil || intVal < -1 {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: "invalid integer limit (expected int >= -1)",
		}
	}

	i.IsSet = true
	i.Value = intVal

	return nil
}

func (i *IntegerLimit) IsValidValue(val string) error {
	return i.UnmarshalFlag(val)
}
