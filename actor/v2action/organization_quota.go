package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type OrganizationQuota ccv2.OrganizationQuota

func (actor Actor) GetOrganizationQuota(guid string) (OrganizationQuota, Warnings, error) {
	orgQuota, warnings, err := actor.CloudControllerClient.GetOrganizationQuota(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return OrganizationQuota{}, Warnings(warnings), actionerror.OrganizationQuotaNotFoundError{GUID: guid}
	}

	return OrganizationQuota(orgQuota), Warnings(warnings), err
}
