package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type SpaceQuota ccv2.SpaceQuota

func (actor Actor) GetSpaceQuota(guid string) (SpaceQuota, Warnings, error) {
	spaceQuota, warnings, err := actor.CloudControllerClient.GetSpaceQuotaDefinition(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return SpaceQuota{}, Warnings(warnings), actionerror.SpaceQuotaNotFoundError{GUID: guid}
	}

	return SpaceQuota(spaceQuota), Warnings(warnings), err
}

// GetSpaceQuotaByName finds the quota by name and returns an error if not found
func (actor Actor) GetSpaceQuotaByName(quotaName, orgGUID string) (SpaceQuota, Warnings, error) {
	quotas, warnings, err := actor.CloudControllerClient.GetSpaceQuotas(orgGUID)

	if err != nil {
		return SpaceQuota{}, Warnings(warnings), err
	}

	for _, quota := range quotas {
		if quota.Name == quotaName {
			return SpaceQuota(quota), Warnings(warnings), nil
		}
	}

	return SpaceQuota{}, Warnings(warnings), actionerror.QuotaNotFoundForNameError{Name: quotaName}
}

// SetSpaceQuota sets the space quota for the corresponding space
func (actor Actor) SetSpaceQuota(spaceGUID, quotaGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.SetSpaceQuota(spaceGUID, quotaGUID)
	return Warnings(warnings), err
}
