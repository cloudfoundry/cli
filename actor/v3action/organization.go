package v3action

import (
	"net/url"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Organization represents a V3 actor organization.
type Organization ccv3.Organization

// GetOrganizationByName returns the organization with the given name.
func (actor Actor) GetOrganizationByName(name string) (Organization, Warnings, error) {
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(url.Values{
		ccv3.NameFilter: []string{name},
	})
	if err != nil {
		return Organization{}, Warnings(warnings), err
	}

	if len(orgs) == 0 {
		return Organization{}, Warnings(warnings), actionerror.OrganizationNotFoundError{Name: name}
	}

	return Organization(orgs[0]), Warnings(warnings), nil
}
