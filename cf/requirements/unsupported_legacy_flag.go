package requirements

import (
	"errors"
	"strings"

	. "code.cloudfoundry.org/cli/v7/cf/i18n"
)

type UnsupportedLegacyFlagRequirement struct {
	flags []string
}

func NewUnsupportedLegacyFlagRequirement(flags []string) UnsupportedLegacyFlagRequirement {
	return UnsupportedLegacyFlagRequirement{
		flags: flags,
	}
}

func (r UnsupportedLegacyFlagRequirement) Execute() error {
	return errors.New(T(
		"The following flags cannot be used with deprecated usage: {{.flags}}",
		map[string]interface{}{
			"flags": strings.Join(r.flags, ", "),
		}))
}
