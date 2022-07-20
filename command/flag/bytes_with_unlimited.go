package flag

import (
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/types"
	flags "github.com/jessevdk/go-flags"
)

type BytesWithUnlimited types.NullInt

func (m *BytesWithUnlimited) UnmarshalFlag(val string) error {
	if val == "" {
		return nil
	}

	if val == "-1" {
		m.Value = -1
		m.IsSet = true
		return nil
	}

	size, err := ConvertToBytes(val)
	if err != nil {
		return err
	}

	m.Value = int(size)
	m.IsSet = true

	return nil
}

func (m *BytesWithUnlimited) IsValidValue(val string) error {
	return m.UnmarshalFlag(val)
}

func ConvertToBytes(val string) (uint64, error) {
	size, err := bytefmt.ToBytes(val)

	if err != nil || strings.Contains(strings.ToLower(val), ".") {
		return size, &flags.Error{
			Type:    flags.ErrRequired,
			Message: `Byte quantity must be an integer with a unit of measurement like B, K, KB, M, MB, G, or GB`,
		}
	}
	return size, nil
}
