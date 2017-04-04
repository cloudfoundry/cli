package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type SpaceQuota ccv2.SpaceQuota

type SpaceQuotaNotFoundError struct {
	GUID string
}

func (e SpaceQuotaNotFoundError) Error() string {
	return fmt.Sprintf("Space quota with GUID '%s' not found.", e.GUID)
}

func (actor Actor) GetSpaceQuota(guid string) (SpaceQuota, Warnings, error) {
	spaceQuota, warnings, err := actor.CloudControllerClient.GetSpaceQuota(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return SpaceQuota{}, Warnings(warnings), SpaceQuotaNotFoundError{GUID: guid}
	}

	return SpaceQuota(spaceQuota), Warnings(warnings), err
}
