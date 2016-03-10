package requirements

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter -o fakes/fake_targeted_org_requirement.go . TargetedOrgRequirement
type TargetedOrgRequirement interface {
	Requirement
	GetOrganizationFields() models.OrganizationFields
}

type targetedOrgApiRequirement struct {
	config core_config.Reader
}

func NewTargetedOrgRequirement(config core_config.Reader) TargetedOrgRequirement {
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
