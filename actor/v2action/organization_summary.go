package v2action

import "sort"

type OrganizationSummary struct {
	Organization
	QuotaName   string
	DomainNames []string
	SpaceNames  []string
}

func (actor Actor) GetOrganizationSummaryByName(orgName string) (OrganizationSummary, Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}

	orgSummary := OrganizationSummary{
		Organization: org,
	}

	domains, warnings, err := actor.GetOrganizationDomains(org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}

	for _, domain := range domains {
		orgSummary.DomainNames = append(orgSummary.DomainNames, domain.Name)
	}
	sort.Strings(orgSummary.DomainNames)

	quota, warnings, err := actor.GetOrganizationQuota(org.QuotaDefinitionGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}
	orgSummary.QuotaName = quota.Name

	spaces, warnings, err := actor.GetOrganizationSpaces(org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return OrganizationSummary{}, allWarnings, err
	}

	for _, space := range spaces {
		orgSummary.SpaceNames = append(orgSummary.SpaceNames, space.Name)
	}
	sort.Strings(orgSummary.SpaceNames)

	return orgSummary, allWarnings, nil
}
