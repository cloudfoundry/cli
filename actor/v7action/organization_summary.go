package v7action

import (
	"sort"

	"code.cloudfoundry.org/cli/resources"
)

type OrganizationSummary struct {
	resources.Organization
	DomainNames []string
	QuotaName   string
	SpaceNames  []string

	// DefaultIsolationSegmentGUID is the unique identifier of the isolation
	// segment this organization is tagged with.
	DefaultIsolationSegmentGUID string
}

func (actor Actor) GetOrganizationSummaryByName(orgName string) (OrganizationSummary, Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}

	domains, warnings, err := actor.GetOrganizationDomains(org.GUID, "")
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}

	quota, ccv3Warnings, err := actor.CloudControllerClient.GetOrganizationQuota(org.QuotaGUID)
	allWarnings = append(allWarnings, ccv3Warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}

	spaces, warnings, err := actor.GetOrganizationSpaces(org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}

	isoSegGUID, warnings, err := actor.GetOrganizationDefaultIsolationSegment(org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}

	domainNames := []string{}
	for _, domain := range domains {
		domainNames = append(domainNames, domain.Name)
	}

	spaceNames := []string{}
	for _, space := range spaces {
		spaceNames = append(spaceNames, space.Name)
	}

	sort.Strings(domainNames)
	sort.Strings(spaceNames)

	organizationSummary := OrganizationSummary{
		Organization:                org,
		DomainNames:                 domainNames,
		QuotaName:                   quota.Name,
		SpaceNames:                  spaceNames,
		DefaultIsolationSegmentGUID: isoSegGUID,
	}

	return organizationSummary, allWarnings, nil
}
