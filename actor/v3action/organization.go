package v3action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/lookuptable"
)

// Organization represents a V3 actor organization.
type Organization resources.Organization

// GetOrganizationByName returns the organization with the given name.
func (actor Actor) GetOrganizationByName(name string) (Organization, Warnings, error) {
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{name}},
	)
	if err != nil {
		return Organization{}, Warnings(warnings), err
	}

	if len(orgs) == 0 {
		return Organization{}, Warnings(warnings), actionerror.OrganizationNotFoundError{Name: name}
	}

	return Organization(orgs[0]), Warnings(warnings), nil
}

func (actor Actor) GetOrganizationsByGUIDs(guids ...string) ([]Organization, Warnings, error) {
	queries := []ccv3.Query{}
	if len(guids) > 0 {
		queries = []ccv3.Query{ccv3.Query{Key: ccv3.GUIDFilter, Values: guids}}
	}

	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(queries...)
	if err != nil {
		return []Organization{}, Warnings(warnings), err
	}

	guidToOrg := lookuptable.OrgFromGUID(orgs)

	filteredOrgs := make([]resources.Organization, 0)
	for _, guid := range guids {
		filteredOrgs = append(filteredOrgs, guidToOrg[guid])
	}
	orgs = filteredOrgs

	return convertCCToActorOrganizations(orgs), Warnings(warnings), nil
}

func (actor Actor) GetOrganizations() ([]Organization, Warnings, error) {
	orderBy := ccv3.Query{
		Key:    "order_by",
		Values: []string{"name"},
	}
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(orderBy)
	if err != nil {
		return []Organization{}, Warnings(warnings), err
	}
	return convertCCToActorOrganizations(orgs), Warnings(warnings), nil
}

func convertCCToActorOrganizations(v3orgs []resources.Organization) []Organization {
	orgs := make([]Organization, len(v3orgs))
	for i := range v3orgs {
		orgs[i] = Organization(v3orgs[i])
	}
	return orgs
}
