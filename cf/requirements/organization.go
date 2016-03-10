package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/models"
)

//go:generate counterfeiter -o fakes/fake_organization_requirement.go . OrganizationRequirement
type OrganizationRequirement interface {
	Requirement
	SetOrganizationName(string)
	GetOrganization() models.Organization
}

type organizationApiRequirement struct {
	name    string
	orgRepo organizations.OrganizationRepository
	org     models.Organization
}

func NewOrganizationRequirement(name string, sR organizations.OrganizationRepository) *organizationApiRequirement {
	req := &organizationApiRequirement{}
	req.name = name
	req.orgRepo = sR
	return req
}

func (req *organizationApiRequirement) Execute() error {
	var apiErr error
	req.org, apiErr = req.orgRepo.FindByName(req.name)

	if apiErr != nil {
		return apiErr
	}

	return nil
}

func (req *organizationApiRequirement) SetOrganizationName(name string) {
	req.name = name
}

func (req *organizationApiRequirement) GetOrganization() models.Organization {
	return req.org
}
