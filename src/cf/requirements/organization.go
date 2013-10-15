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

type organizationApiRequirement struct {
	name    string
	ui      terminal.UI
	orgRepo api.OrganizationRepository
	org     cf.Organization
}

func newOrganizationRequirement(name string, ui terminal.UI, sR api.OrganizationRepository) (req *organizationApiRequirement) {
	req = new(organizationApiRequirement)
	req.name = name
	req.ui = ui
	req.orgRepo = sR
	return
}

func (req *organizationApiRequirement) Execute() (success bool) {
	var apiResponse net.ApiResponse
	req.org, apiResponse = req.orgRepo.FindByName(req.name)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *organizationApiRequirement) GetOrganization() cf.Organization {
	return req.org
}
