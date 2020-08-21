package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
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

	return handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			_, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, targetedSpaceGUID)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			if sharedToDetails.OrgName.IsSet {
				orgName := sharedToDetails.OrgName.Value

				var organizations []resources.Organization
				organizations, warnings, err = actor.CloudControllerClient.GetOrganizations(
					[]ccv3.Query{ccv3.Query{Key: ccv3.NameFilter, Values: []string{orgName}}}...,
				)
				if err == nil && len(organizations) == 0 {
					err = actionerror.OrganizationNotFoundError{Name: orgName}
				}
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			var spaceWarnings Warnings
			_, spaceWarnings, err = actor.GetSpaceByNameAndOrganization(sharedToDetails.SpaceName, targetedOrgGUID)
			warnings = ccv3.Warnings(spaceWarnings)
			return
		},
	))

	return Warnings{}, nil
}
