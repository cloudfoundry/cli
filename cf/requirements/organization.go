package requirements

import (
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/models"
)

//go:generate counterfeiter . OrganizationRequirement

type OrganizationRequirement interface {
	Requirement
	SetOrganizationName(string)
	GetOrganization() models.Organization
}

type organizationAPIRequirement struct {
	name    string
	orgRepo organizations.OrganizationRepository
	org     models.Organization
}

func NewOrganizationRequirement(name string, sR organizations.OrganizationRepository) *organizationAPIRequirement {
	req := &organizationAPIRequirement{}
	req.name = name
	req.orgRepo = sR
	return req
}

func (req *organizationAPIRequirement) Execute() error {
	var apiErr error
	req.org, apiErr = req.orgRepo.FindByName(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *organizationAPIRequirement) SetOrganizationName(name string) {
	req.name = name
}

func (req *organizationAPIRequirement) GetOrganization() models.Organization {
	return req.org
}
