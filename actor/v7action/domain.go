package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
)

type Domain ccv3.Domain

type SharedOrgs ccv3.SharedOrgs

func (domain Domain) Shared() bool {
	return domain.OrganizationGUID == ""
}

func (actor Actor) CheckRoute(domainName string, hostname string, path string) (bool, Warnings, error) {
	var allWarnings Warnings

	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return false, allWarnings, err
	}

	matches, checkRouteWarnings, err := actor.CloudControllerClient.CheckRoute(domain.GUID, hostname, path)
	allWarnings = append(allWarnings, checkRouteWarnings...)

	return matches, allWarnings, err
}

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
		OrganizationGUID: organization.GUID,
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	return allWarnings, err
}

func (actor Actor) DeleteDomain(domain Domain) (Warnings, error) {
	allWarnings := Warnings{}

	jobURL, apiWarnings, err := actor.CloudControllerClient.DeleteDomain(domain.GUID)

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if err != nil {
		return allWarnings, err
	}

	pollJobWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, Warnings(pollJobWarnings)...)

	return allWarnings, err
}

func (actor Actor) GetOrganizationDomains(orgGuid string) ([]Domain, Warnings, error) {
	ccv3Domains, warnings, err := actor.CloudControllerClient.GetOrganizationDomains(orgGuid)

	if err != nil {
		return nil, Warnings(warnings), err
	}

	var domains []Domain
	for _, domain := range ccv3Domains {
		domains = append(domains, Domain(domain))
	}

	return domains, Warnings(warnings), nil
}

func (actor Actor) GetDomain(domainGUID string) (Domain, Warnings, error) {
	domain, warnings, err := actor.CloudControllerClient.GetDomain(domainGUID)

	if err != nil {
		return Domain{}, Warnings(warnings), err
	}

	return Domain(domain), Warnings(warnings), nil
}

func (actor Actor) GetDomainByName(domainName string) (Domain, Warnings, error) {
	domains, warnings, err := actor.CloudControllerClient.GetDomains(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{domainName}},
	)

	if err != nil {
		return Domain{}, Warnings(warnings), err
	}

	if len(domains) == 0 {
		return Domain{}, Warnings(warnings), actionerror.DomainNotFoundError{Name: domainName}
	}

	return Domain(domains[0]), Warnings(warnings), nil
}

func (actor Actor) SharePrivateDomain(domainName string, orgName string) (Warnings, error) {
	orgGUID, domainGUID, warnings, err := actor.GetDomainAndOrgGUIDsByName(domainName, orgName)

	if err != nil {
		return warnings, err
	}

	apiWarnings, err := actor.CloudControllerClient.SharePrivateDomainToOrgs(
		domainGUID,
		ccv3.SharedOrgs{GUIDs: []string{orgGUID}},
	)

	allWarnings := append(warnings, Warnings(apiWarnings)...)

	return allWarnings, err
}

func (actor Actor) UnsharePrivateDomain(domainName string, orgName string) (Warnings, error) {
	orgGUID, domainGUID, warnings, err := actor.GetDomainAndOrgGUIDsByName(domainName, orgName)

	if err != nil {
		return warnings, err
	}

	apiWarnings, err := actor.CloudControllerClient.UnsharePrivateDomainFromOrg(
		domainGUID,
		orgGUID,
	)

	allWarnings := append(warnings, Warnings(apiWarnings)...)

	return allWarnings, err
}

func (actor Actor) GetDomainAndOrgGUIDsByName(domainName string, orgName string) (string, string, Warnings, error) {
	org, getOrgWarnings, err := actor.GetOrganizationByName(orgName)

	if err != nil {
		return "", "", getOrgWarnings, err
	}

	domain, getDomainWarnings, err := actor.GetDomainByName(domainName)
	allWarnings := append(getOrgWarnings, getDomainWarnings...)

	if err != nil {
		return "", "", allWarnings, err
	}

	return org.GUID, domain.GUID, allWarnings, nil
}
