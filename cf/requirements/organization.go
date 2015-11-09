package requirements

import (
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter -o fakes/fake_organization_requirement.go . OrganizationRequirement
type OrganizationRequirement interface {
	Requirement
	SetOrganizationName(string)
	GetOrganization() models.Organization
}

type organizationApiRequirement struct {
	name    string
	ui      terminal.UI
	orgRepo organizations.OrganizationRepository
	org     models.Organization
}

func NewOrganizationRequirement(name string, ui terminal.UI, sR organizations.OrganizationRepository) *organizationApiRequirement {
	req := &organizationApiRequirement{}
	req.name = name
	req.ui = ui
	req.orgRepo = sR
	return req
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

func (req *organizationApiRequirement) SetOrganizationName(name string) {
	req.name = name
}

func (req *organizationApiRequirement) GetOrganization() models.Organization {
	return req.org
}
