package flag

import (
	"strconv"

	flags "github.com/jessevdk/go-flags"
)

type PositiveInteger struct {
	Value int64
}

func (posInt *PositiveInteger) UnmarshalFlag(rawValue string) error {
	value, err := strconv.ParseInt(rawValue, 10, 0)
	if err != nil {
		return err
	}

	if value < 1 {
		return &flags.Error{
			Type:    flags.ErrMarshal,
			Message: `Value must be greater than or equal to 1.`,
		}
	}

	posInt.Value = value
	return nil
}
