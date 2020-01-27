package flag

import (
	"strings"

	"code.cloudfoundry.org/cli/types"

	"code.cloudfoundry.org/bytefmt"
	flags "github.com/jessevdk/go-flags"
)

const (
	AllowedUnits = "mg"
)

type Megabytes struct {
	types.NullUint64
}

func (m *Megabytes) UnmarshalFlag(val string) error {
	if val == "" {
		return nil
	}

	size, err := ConvertToMb(val)
	if err != nil {
		return err
	}

	m.Value = size
	m.IsSet = true

	return nil
}

func ConvertToMb(val string) (uint64, error) {
	size, err := bytefmt.ToMegabytes(val)

	if err != nil ||
		!strings.ContainsAny(strings.ToLower(val), AllowedUnits) ||
		strings.Contains(strings.ToLower(val), ".") {
		return size, &flags.Error{
			Type:    flags.ErrRequired,
			Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
		}
	}
	return size, nil
}
