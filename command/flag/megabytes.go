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

// When setting a max memory limit, allows for -1 as unlimited
type MegabytesUnlimited types.NullInt

func (m *Megabytes) UnmarshalFlag(val string) error {
	if val == "" {
		return nil
	}

	size, err := bytefmt.ToMegabytes(val)

	if err != nil ||
		!strings.ContainsAny(strings.ToLower(val), AllowedUnits) ||
		strings.Contains(strings.ToLower(val), ".") {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
		}
	}

	m.Value = size
	m.IsSet = true

	return nil
}

func (m *MegabytesUnlimited) UnmarshalFlag(val string) error {
	if val == "" {
		return nil
	}

	if val == "-1" {
		m.Value = -1
		m.IsSet = true
		return nil
	}

	size, err := bytefmt.ToMegabytes(val)

	if err != nil ||
		!strings.ContainsAny(strings.ToLower(val), AllowedUnits) ||
		strings.Contains(strings.ToLower(val), ".") {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
		}
	}

	m.Value = int(size)
	m.IsSet = true

	return nil
}
