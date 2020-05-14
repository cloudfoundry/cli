package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) GetOrganizations(labelSelector string) ([]resources.Organization, Warnings, error) {
	queries := []ccv3.Query{
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
	}
	if len(labelSelector) > 0 {
		queries = append(queries, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(queries...)
	if err != nil {
		return []resources.Organization{}, Warnings(warnings), err
	}

	return orgs, Warnings(warnings), nil
}

// GetOrganizationByGUID returns the organization with the given guid.
func (actor Actor) GetOrganizationByGUID(orgGUID string) (resources.Organization, Warnings, error) {
	ccOrg, warnings, err := actor.CloudControllerClient.GetOrganization(orgGUID)
	if err != nil {
		return resources.Organization{}, Warnings(warnings), err
	}

	return resources.Organization(ccOrg), Warnings(warnings), err
}

// GetOrganizationByName returns the organization with the given name.
func (actor Actor) GetOrganizationByName(name string) (resources.Organization, Warnings, error) {
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{name}},
	)
	if err != nil {
		return resources.Organization{}, Warnings(warnings), err
	}

	if len(orgs) == 0 {
		return resources.Organization{}, Warnings(warnings), actionerror.OrganizationNotFoundError{Name: name}
	}

	return resources.Organization(orgs[0]), Warnings(warnings), nil
}

// CreateOrganization creates a new organization with the given name
func (actor Actor) CreateOrganization(orgName string) (resources.Organization, Warnings, error) {
	allWarnings := Warnings{}

	organization, apiWarnings, err := actor.CloudControllerClient.CreateOrganization(orgName)
	allWarnings = append(allWarnings, apiWarnings...)

	return organization, allWarnings, err
}

// updateOrganization updates the name and/or labels of an organization
func (actor Actor) updateOrganization(org resources.Organization) (resources.Organization, Warnings, error) {
	updatedOrg, warnings, err := actor.CloudControllerClient.UpdateOrganization(org)
	if err != nil {
		return resources.Organization{}, Warnings(warnings), err
	}

	return updatedOrg, Warnings(warnings), nil
}

func (actor Actor) RenameOrganization(oldOrgName, newOrgName string) (resources.Organization, Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(oldOrgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Organization{}, allWarnings, err
	}

	org.Name = newOrgName
	org, warnings, err = actor.updateOrganization(org)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Organization{}, allWarnings, err
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
