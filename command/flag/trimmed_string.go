package flag

import (
	"strings"
)

type TrimmedString string

func (t *TrimmedString) UnmarshalFlag(value string) error {
	*t = TrimmedString(strings.TrimSpace(value))
	return nil
}
