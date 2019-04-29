package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Organization represents a V3 actor organization.
type Organization struct {
	// GUID is the unique organization identifier.
	GUID string
	// Name is the name of the organization.
	Name string

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata
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

	return actor.convertCCToActorOrganization(orgs[0]), Warnings(warnings), nil
}

// UpdateOrganization updates the labels of an organization
func (actor Actor) UpdateOrganization(org Organization) (Organization, Warnings, error) {
	ccOrg := ccv3.Organization{
		GUID:     org.GUID,
		Name:     org.Name,
		Metadata: (*ccv3.Metadata)(org.Metadata),
	}

	updatedOrg, warnings, err := actor.CloudControllerClient.UpdateOrganization(ccOrg)
	if err != nil {
		return Organization{}, Warnings(warnings), err
	}

	return actor.convertCCToActorOrganization(updatedOrg), Warnings(warnings), nil
}

func (actor Actor) convertCCToActorOrganization(org ccv3.Organization) Organization {
	return Organization{
		GUID:     org.GUID,
		Name:     org.Name,
		Metadata: (*Metadata)(org.Metadata),
	}
}

func (actor Actor) GetDefaultDomain(orgGUID string) (Domain, Warnings, error) {
	domain, warnings, err := actor.CloudControllerClient.GetDefaultDomain(orgGUID)

	return Domain(domain), Warnings(warnings), err
}
