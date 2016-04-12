package requirements

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter . TargetedOrgRequirement

type TargetedOrgRequirement interface {
	Requirement
	GetOrganizationFields() models.OrganizationFields
}

type targetedOrgApiRequirement struct {
	config coreconfig.Reader
}

func NewTargetedOrgRequirement(config coreconfig.Reader) TargetedOrgRequirement {
	return targetedOrgApiRequirement{config}
}

func (req targetedOrgApiRequirement) Execute() error {
	if !req.config.HasOrganization() {
		message := fmt.Sprintf(T("No org targeted, use '{{.Command}}' to target an org.", map[string]interface{}{"Command": terminal.CommandColor(cf.Name + " target -o ORG")}))
		return errors.New(message)
	}

	return nil
}

func (req targetedOrgApiRequirement) GetOrganizationFields() (org models.OrganizationFields) {
	return req.config.OrganizationFields()
}
