package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

type Domain ccv3.Domain

func (actor Actor) CreateSharedDomain(domainName string, internal bool) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.CreateDomain(ccv3.Domain{
		Name:     domainName,
		Internal: types.NullBool{IsSet: true, Value: internal},
	})
	return Warnings(warnings), err
}

func (actor Actor) CreatePrivateDomain(domainName string, orgName string) (Warnings, error) {
	allWarnings := Warnings{}
	organization, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}
	_, apiWarnings, err := actor.CloudControllerClient.CreateDomain(ccv3.Domain{
		Name:             domainName,
		OrganizationGuid: organization.GUID,
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, err
}

// TODO: filter by targeted org
func (actor Actor) GetDomains() ([]Domain, Warnings, error) {
	ccv3Domains, warnings, err := actor.CloudControllerClient.GetDomains()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var domains []Domain
	for _, domain := range ccv3Domains {
		domains = append(domains, Domain(domain))
	}
	return domains, Warnings(warnings), nil
}
