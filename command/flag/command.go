package flag

import (
	"code.cloudfoundry.org/cli/types"
)

type Command struct {
	types.FilteredString
}

func (b *Command) UnmarshalFlag(val string) error {
	b.ParseValue(val)
	return nil
}
