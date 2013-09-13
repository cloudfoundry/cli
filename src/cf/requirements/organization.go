package requirements

import (
	"cf"
	"cf/api"
	"cf/terminal"
)

type OrganizationRequirement interface {
	Requirement
	GetOrganization() cf.Organization
}

type OrganizationApiRequirement struct {
	name    string
	ui      terminal.UI
	orgRepo api.OrganizationRepository
	org     cf.Organization
}

func NewOrganizationRequirement(name string, ui terminal.UI, sR api.OrganizationRepository) (req *OrganizationApiRequirement) {
	req = new(OrganizationApiRequirement)
	req.name = name
	req.ui = ui
	req.orgRepo = sR
	return
}

func (req *OrganizationApiRequirement) Execute() (success bool) {
	var apiErr *api.ApiError
	req.org, apiErr = req.orgRepo.FindByName(req.name)

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *OrganizationApiRequirement) GetOrganization() cf.Organization {
	return req.org
}
