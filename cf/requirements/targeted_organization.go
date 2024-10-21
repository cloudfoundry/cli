package requirements

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/v9/cf"
	"code.cloudfoundry.org/cli/v9/cf/configuration/coreconfig"
	. "code.cloudfoundry.org/cli/v9/cf/i18n"
	"code.cloudfoundry.org/cli/v9/cf/models"
	"code.cloudfoundry.org/cli/v9/cf/terminal"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . TargetedOrgRequirement

type TargetedOrgRequirement interface {
	Requirement
	GetOrganizationFields() models.OrganizationFields
}

type targetedOrgAPIRequirement struct {
	config coreconfig.Reader
}

func NewTargetedOrgRequirement(config coreconfig.Reader) TargetedOrgRequirement {
	return targetedOrgAPIRequirement{config}
}

func (req targetedOrgAPIRequirement) Execute() error {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf(T("No org targeted, use '{{.Command}}' to target an org.", map[string]interface{}{"Command": terminal.CommandColor(cf.Name + " target -o ORG")}))
		return errors.New(message)
	}

	return nil
}

func (req targetedOrgAPIRequirement) GetOrganizationFields() (org models.OrganizationFields) {
	return req.config.OrganizationFields()
}
