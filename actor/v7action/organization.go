package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Organization represents a V7 actor organization.
type Organization ccv3.Organization

func (actor Actor) GetOrganizations(labelSelector string) ([]Organization, Warnings, error) {
	queries := []ccv3.Query{
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
	}
	if len(labelSelector) > 0 {
		queries = append(queries, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}
	ccOrgs, warnings, err := actor.CloudControllerClient.GetOrganizations(queries...)
	if err != nil {
		return []Organization{}, Warnings(warnings), err
	}

	orgs := make([]Organization, len(ccOrgs))
	for i, ccOrg := range ccOrgs {
		orgs[i] = Organization(ccOrg)
	}

	return orgs, Warnings(warnings), nil
}

// GetOrganizationByGUID returns the organization with the given guid.
func (actor Actor) GetOrganizationByGUID(orgGUID string) (Organization, Warnings, error) {
	ccOrg, warnings, err := actor.CloudControllerClient.GetOrganization(orgGUID)
	if err != nil {
		return Organization{}, Warnings(warnings), err
	}

	return Organization(ccOrg), Warnings(warnings), err
}

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

// CreateOrganization creates a new organization with the given name
func (actor Actor) CreateOrganization(orgName string) (Organization, Warnings, error) {
	allWarnings := Warnings{}

	organization, apiWarnings, err := actor.CloudControllerClient.CreateOrganization(orgName)
	allWarnings = append(allWarnings, apiWarnings...)

	return Organization(organization), allWarnings, err
}

// updateOrganization updates the name and/or labels of an organization
func (actor Actor) updateOrganization(org Organization) (Organization, Warnings, error) {
	ccOrg := ccv3.Organization{
		GUID:     org.GUID,
		Name:     org.Name,
		Metadata: org.Metadata,
	}

	updatedOrg, warnings, err := actor.CloudControllerClient.UpdateOrganization(ccOrg)
	if err != nil {
		return Organization{}, Warnings(warnings), err
	}

	return Organization(updatedOrg), Warnings(warnings), nil
}

func (actor Actor) RenameOrganization(oldOrgName, newOrgName string) (Organization, Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(oldOrgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Organization{}, allWarnings, err
	}

	org.Name = newOrgName
	org, warnings, err = actor.updateOrganization(org)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Organization{}, allWarnings, err
	}
	return org, allWarnings, nil
}

func (actor Actor) DeleteOrganization(name string) (Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(name)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	jobURL, deleteWarnings, err := actor.CloudControllerClient.DeleteOrganization(org.GUID)
	allWarnings = append(allWarnings, Warnings(deleteWarnings)...)
	if err != nil {
		return allWarnings, err
	}

	ccWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, Warnings(ccWarnings)...)

	return allWarnings, err
}

func (actor Actor) GetDefaultDomain(orgGUID string) (Domain, Warnings, error) {
	domain, warnings, err := actor.CloudControllerClient.GetDefaultDomain(orgGUID)

	return Domain(domain), Warnings(warnings), err
}
