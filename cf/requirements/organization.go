package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type OrganizationRequirement interface {
	Requirement
	GetOrganization() models.Organization
}

type organizationApiRequirement struct {
	name    string
	ui      terminal.UI
	orgRepo organizations.OrganizationRepository
	org     models.Organization
}

func NewOrganizationRequirement(name string, ui terminal.UI, sR organizations.OrganizationRepository) (req *organizationApiRequirement) {
	req = new(organizationApiRequirement)
	req.name = name
	req.ui = ui
	req.orgRepo = sR
	return
}

func (req *organizationApiRequirement) Execute() (success bool) {
	var apiErr error
	req.org, apiErr = req.orgRepo.FindByName(req.name)

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *organizationApiRequirement) GetOrganization() models.Organization {
	return req.org
}
