package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
)

type OrganizationQuota ccv2.OrganizationQuota

type OrganizationQuotaNotFoundError struct {
	GUID string
}

func (e OrganizationQuotaNotFoundError) Error() string {
	return fmt.Sprintf("Organization quota with GUID '%s' not found.", e.GUID)
}

func (actor Actor) GetOrganizationQuota(guid string) (OrganizationQuota, Warnings, error) {
	orgQuota, warnings, err := actor.CloudControllerClient.GetOrganizationQuota(guid)

	if _, ok := err.(cloudcontroller.ResourceNotFoundError); ok {
		return OrganizationQuota{}, Warnings(warnings), OrganizationQuotaNotFoundError{GUID: guid}
	}

	return OrganizationQuota(orgQuota), Warnings(warnings), err
}
