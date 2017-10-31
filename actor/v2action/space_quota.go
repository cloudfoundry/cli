package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type SpaceQuota ccv2.SpaceQuota

func (actor Actor) GetSpaceQuota(guid string) (SpaceQuota, Warnings, error) {
	spaceQuota, warnings, err := actor.CloudControllerClient.GetSpaceQuota(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return SpaceQuota{}, Warnings(warnings), actionerror.SpaceQuotaNotFoundError{GUID: guid}
	}

	return SpaceQuota(spaceQuota), Warnings(warnings), err
}
