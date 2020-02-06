package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type OrganizationQuota ccv2.OrganizationQuota

func (actor Actor) GetOrganizationQuota(guid string) (OrganizationQuota, Warnings, error) {
	orgQuota, warnings, err := actor.CloudControllerClient.GetOrganizationQuota(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return OrganizationQuota{}, Warnings(warnings), actionerror.OrganizationQuotaNotFoundError{GUID: guid}
	}

	return OrganizationQuota(orgQuota), Warnings(warnings), err
}

func (actor Actor) GetOrganizationQuotaByName(quotaName string) (OrganizationQuota, Warnings, error) {
	orgQuotas, warnings, err := actor.CloudControllerClient.GetOrganizationQuotas(ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{quotaName},
	})
	if err != nil {
		return OrganizationQuota{}, Warnings(warnings), err
	}

	if len(orgQuotas) > 1 {
		var GUIDs []string
		for _, orgQuota := range orgQuotas {
			GUIDs = append(GUIDs, orgQuota.GUID)
		}
		return OrganizationQuota{}, Warnings(warnings), actionerror.MultipleOrganizationQuotasFoundForNameError{Name: quotaName, GUIDs: GUIDs}
	} else if len(orgQuotas) == 0 {
		return OrganizationQuota{}, Warnings(warnings), actionerror.QuotaNotFoundForNameError{Name: quotaName}
	}

	return OrganizationQuota(orgQuotas[0]), Warnings(warnings), nil
}
