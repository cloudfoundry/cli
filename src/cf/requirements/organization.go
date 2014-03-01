package requirements

import (
	"cf/api"
	"cf/errors"
	"cf/models"
	"cf/terminal"
)

type OrganizationRequirement interface {
	Requirement
	GetOrganization() models.Organization
}

type organizationApiRequirement struct {
	name    string
	ui      terminal.UI
	orgRepo api.OrganizationRepository
	org     models.Organization
}

func NewOrganizationRequirement(name string, ui terminal.UI, sR api.OrganizationRepository) (req *organizationApiRequirement) {
	req = new(organizationApiRequirement)
	req.name = name
	req.ui = ui
	req.orgRepo = sR
	return
}

func (req *organizationApiRequirement) Execute() (success bool) {
	var apiResponse errors.Error
	req.org, apiResponse = req.orgRepo.FindByName(req.name)

	if apiResponse != nil {
		req.ui.Failed(apiResponse.Error())
		return false
	}

	return true
}

func (req *organizationApiRequirement) GetOrganization() models.Organization {
	return req.org
}
