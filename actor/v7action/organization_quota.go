package v7action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

type OrganizationQuota ccv3.OrgQuota

func (actor Actor) GetOrganizationQuotas() ([]OrganizationQuota, Warnings, error) {
	ccv3OrgQuotas, warnings, err := actor.CloudControllerClient.GetOrganizationQuotas()
	if err != nil {
		return []OrganizationQuota{}, Warnings(warnings), err
	}

	var orgQuotas []OrganizationQuota
	for _, quota := range ccv3OrgQuotas {
		orgQuotas = append(orgQuotas, OrganizationQuota(quota))
	}

	return orgQuotas, Warnings(warnings), nil
}
