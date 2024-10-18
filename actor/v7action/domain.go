package v7action

import (
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/types"
)

type SharedOrgs ccv3.SharedOrgs

func (actor Actor) CheckRoute(domainName string, hostname string, path string, port int) (bool, Warnings, error) {
	var allWarnings Warnings

	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return false, allWarnings, err
	}

	matches, checkRouteWarnings, err := actor.CloudControllerClient.CheckRoute(domain.GUID, hostname, path, port)
	allWarnings = append(allWarnings, checkRouteWarnings...)

	return matches, allWarnings, err
}

func (actor Actor) CreateSharedDomain(domainName string, internal bool, routerGroupName string) (Warnings, error) {
	allWarnings := Warnings{}
	routerGroupGUID := ""

	if routerGroupName != "" {
		routerGroup, err := actor.GetRouterGroupByName(routerGroupName)
		if err != nil {
			return allWarnings, err
		}

		routerGroupGUID = routerGroup.GUID
	}

	_, warnings, err := actor.CloudControllerClient.CreateDomain(resources.Domain{
		Name:        domainName,
		Internal:    types.NullBool{IsSet: true, Value: internal},
		RouterGroup: routerGroupGUID,
	})
	allWarnings = append(allWarnings, Warnings(warnings)...)

	return allWarnings, err
}

func (actor Actor) CreatePrivateDomain(domainName string, orgName string) (Warnings, error) {
	allWarnings := Warnings{}
	organization, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}
	_, apiWarnings, err := actor.CloudControllerClient.CreateDomain(resources.Domain{
		Name:             domainName,
		OrganizationGUID: organization.GUID,
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	return allWarnings, err
}

func (actor Actor) DeleteDomain(domain resources.Domain) (Warnings, error) {
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

func (actor Actor) GetOrganizationDomains(orgGuid string, labelSelector string) ([]resources.Domain, Warnings, error) {
	keys := []ccv3.Query{}
	if labelSelector != "" {
		keys = append(keys, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}
	ccv3Domains, warnings, err := actor.CloudControllerClient.GetOrganizationDomains(orgGuid, keys...)

	if err != nil {
		return nil, Warnings(warnings), err
	}

	var domains []resources.Domain
	for _, domain := range ccv3Domains {
		domains = append(domains, resources.Domain(domain))
	}

	return domains, Warnings(warnings), nil
}

func (actor Actor) GetDomain(domainGUID string) (resources.Domain, Warnings, error) {
	domain, warnings, err := actor.CloudControllerClient.GetDomain(domainGUID)

	if err != nil {
		return resources.Domain{}, Warnings(warnings), err
	}

	return resources.Domain(domain), Warnings(warnings), nil
}

func (actor Actor) GetDomainByName(domainName string) (resources.Domain, Warnings, error) {
	domain, warnings, err := actor.getDomainByName(domainName)
	return domain, Warnings(warnings), err
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

func (actor Actor) getDomainByName(domainName string) (resources.Domain, ccv3.Warnings, error) {
	domains, warnings, err := actor.CloudControllerClient.GetDomains(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{domainName}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)
	switch {
	case err != nil:
		return resources.Domain{}, warnings, err
	case len(domains) == 0:
		return resources.Domain{}, warnings, actionerror.DomainNotFoundError{Name: domainName}
	default:
		return domains[0], warnings, nil
	}
}
