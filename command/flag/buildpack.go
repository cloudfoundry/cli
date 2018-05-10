package flag

import (
	"code.cloudfoundry.org/cli/types"
)

type Buildpack struct {
	types.FilteredString
}

func (b *Buildpack) UnmarshalFlag(val string) error {
	b.ParseValue(val)
	return nil
}
