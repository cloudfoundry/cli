package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/railway"
)

type ServiceInstanceSharingParams struct {
	SpaceName string
	OrgName   types.OptionalString
}

func (actor Actor) ShareServiceInstanceToSpaceAndOrg(
	serviceInstanceName, targetedSpaceGUID, targetedOrgGUID string,
	sharedToDetails ServiceInstanceSharingParams,
) (Warnings, error) {
	var shareToOrgGUID = targetedOrgGUID

	return handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			_, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, targetedSpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if sharedToDetails.OrgName.IsSet {
				var (
					orgWarnings  Warnings
					organization resources.Organization
				)

				organization, orgWarnings, err = actor.GetOrganizationByName(sharedToDetails.OrgName.Value)
				warnings = ccv3.Warnings(orgWarnings)
				if err == nil {
					shareToOrgGUID = organization.GUID
				}
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			var spaceWarnings Warnings
			_, spaceWarnings, err = actor.GetSpaceByNameAndOrganization(sharedToDetails.SpaceName, shareToOrgGUID)
			warnings = ccv3.Warnings(spaceWarnings)
			return
		},
	))

	return Warnings{}, nil
}
