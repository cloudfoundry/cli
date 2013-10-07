package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
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
	var apiResponse net.ApiResponse
	req.org, apiResponse = req.orgRepo.FindByName(req.name)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *OrganizationApiRequirement) GetOrganization() cf.Organization {
	return req.org
}
