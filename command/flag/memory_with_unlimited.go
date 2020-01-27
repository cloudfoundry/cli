package flag

import (
	"code.cloudfoundry.org/cli/types"
)

type MemoryWithUnlimited types.NullInt

func (m *MemoryWithUnlimited) UnmarshalFlag(val string) error {
	if val == "" {
		return nil
	}

	if val == "-1" {
		m.Value = -1
		m.IsSet = true
		return nil
	}

	size, err := ConvertToMb(val)
	if err != nil {
		return err
	}

	m.Value = int(size)
	m.IsSet = true

	return nil
}

func (m *MemoryWithUnlimited) IsValidValue(val string) error {
	return m.UnmarshalFlag(val)
}
