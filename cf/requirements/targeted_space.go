package requirements

import (
	"errors"

	"code.cloudfoundry.org/cli/v8/cf"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	. "code.cloudfoundry.org/cli/v8/cf/i18n"
	"code.cloudfoundry.org/cli/v8/cf/terminal"
)

type TargetedSpaceRequirement struct {
	config coreconfig.Reader
}

func NewTargetedSpaceRequirement(config coreconfig.Reader) TargetedSpaceRequirement {
	return TargetedSpaceRequirement{config}
}

func (req TargetedSpaceRequirement) Execute() error {
	if !req.config.HasOrganization() {
		message := T("No org and space targeted, use '{{.Command}}' to target an org and space", map[string]interface{}{"Command": terminal.CommandColor(cf.Name + " target -o ORG -s SPACE")})
		return errors.New(message)
	}

	if !req.config.HasSpace() {
		message := T("No space targeted, use '{{.Command}}' to target a space.", map[string]interface{}{"Command": terminal.CommandColor("cf target -s")})
		return errors.New(message)
	}

	return nil
}
