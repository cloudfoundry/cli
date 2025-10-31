package flag

import "code.cloudfoundry.org/cli/v9/types"

type OptionalString types.OptionalString

func (o *OptionalString) UnmarshalFlag(val string) error {
	o.IsSet = true
	o.Value = val

	return nil
}
