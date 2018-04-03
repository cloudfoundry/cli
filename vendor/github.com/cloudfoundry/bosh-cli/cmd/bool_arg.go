package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type BoolArg bool

func (v *BoolArg) UnmarshalFlag(data string) error {
	for _, i := range []string{"true", "yes", "on", "enable"} {
		if i == data {
			*v = true
			return nil
		}
	}

	for _, i := range []string{"false", "no", "off", "disable"} {
		if i == data {
			*v = false
			return nil
		}
	}

	return bosherr.Errorf("Expected boolean variable '%s' to be either 'true' or 'false'", data)
}
