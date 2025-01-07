package flag

import (
	"regexp"
	"strings"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/v8/types"
	flags "github.com/jessevdk/go-flags"
)

type BytesWithUnlimited types.NullInt

var zeroBytes *regexp.Regexp = regexp.MustCompile(`^0[KMGT]?B?$`)
var negativeOneBytes *regexp.Regexp = regexp.MustCompile(`^-1[KMGT]?B?$`)

func (m *BytesWithUnlimited) UnmarshalFlag(val string) error {
	if val == "" {
		return nil
	}

	if negativeOneBytes.MatchString(val) {
		m.Value = -1
		m.IsSet = true
		return nil
	}

	if zeroBytes.MatchString(val) {
		m.Value = 0
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
