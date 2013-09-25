package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/terminal"
	"fmt"
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
	var (
		apiErr *net.ApiError
		found  bool
	)
	req.org, found, apiErr = req.orgRepo.FindByName(req.name)

	if !found {
		req.ui.Failed(fmt.Sprintf("Organization %s could not be found.", terminal.EntityNameColor(req.name)))
		return false
	}

	if apiErr != nil {
		req.ui.Failed(apiErr.Error())
		return false
	}

	return true
}

func (req *OrganizationApiRequirement) GetOrganization() cf.Organization {
	return req.org
}
