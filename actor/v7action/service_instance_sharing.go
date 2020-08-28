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
	var serviceInstance resources.ServiceInstance
	var shareSpace resources.Space

	return handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			var actorWarnings Warnings
			serviceInstance, shareSpace, actorWarnings, err = actor.validateSharingDetails(serviceInstanceName, targetedSpaceGUID, targetedOrgGUID, sharedToDetails)
			warnings = ccv3.Warnings(actorWarnings)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			_, warnings, err = actor.CloudControllerClient.ShareServiceInstanceToSpaces(serviceInstance.GUID, []string{shareSpace.GUID})
			return
		},
	))

	return Warnings{}, nil
}

func (actor Actor) UnshareServiceInstanceFromSpaceAndOrg(
	serviceInstanceName, targetedSpaceGUID, targetedOrgGUID string,
	sharedToDetails ServiceInstanceSharingParams,
) (Warnings, error) {
	return handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			var actorWarnings Warnings
			_, _, actorWarnings, err = actor.validateSharingDetails(serviceInstanceName, targetedSpaceGUID, targetedOrgGUID, sharedToDetails)
			warnings = ccv3.Warnings(actorWarnings)
			return
		},
	))

	return Warnings{}, nil
}

func (actor Actor) validateSharingDetails(
	serviceInstanceName, targetedSpaceGUID, targetedOrgGUID string,
	sharedToDetails ServiceInstanceSharingParams,
) (resources.ServiceInstance, resources.Space, Warnings, error) {
	var serviceInstance resources.ServiceInstance
	var shareSpace resources.Space
	var shareToOrgGUID = targetedOrgGUID

	warnings, err := handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, _, warnings, err = actor.CloudControllerClient.GetServiceInstanceByNameAndSpace(serviceInstanceName, targetedSpaceGUID)
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
				shareToOrgGUID = organization.GUID
			}
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			var spaceWarnings Warnings
			shareSpace, spaceWarnings, err = actor.GetSpaceByNameAndOrganization(sharedToDetails.SpaceName, shareToOrgGUID)
			warnings = ccv3.Warnings(spaceWarnings)
			return
		},
	))

	if err != nil {
		return resources.ServiceInstance{}, resources.Space{}, warnings, err
	}

	return serviceInstance, shareSpace, warnings, nil
}
