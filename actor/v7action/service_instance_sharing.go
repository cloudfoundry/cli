package v7action

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/types"
	"code.cloudfoundry.org/cli/v8/util/railway"
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
			serviceInstance, shareSpace, warnings, err = actor.validateSharingDetails(serviceInstanceName, targetedSpaceGUID, targetedOrgGUID, sharedToDetails)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			_, warnings, err = actor.CloudControllerClient.ShareServiceInstanceToSpaces(serviceInstance.GUID, []string{shareSpace.GUID})
			return
		},
	))
}

func (actor Actor) UnshareServiceInstanceFromSpaceAndOrg(
	serviceInstanceName, targetedSpaceGUID, targetedOrgGUID string,
	sharedToDetails ServiceInstanceSharingParams,
) (Warnings, error) {
	var serviceInstance resources.ServiceInstance
	var unshareSpace resources.Space

	return handleServiceInstanceErrors(railway.Sequentially(
		func() (warnings ccv3.Warnings, err error) {
			serviceInstance, unshareSpace, warnings, err = actor.validateSharingDetails(
				serviceInstanceName,
				targetedSpaceGUID,
				targetedOrgGUID,
				sharedToDetails,
			)
			return
		},
		func() (warnings ccv3.Warnings, err error) {
			warnings, err = actor.CloudControllerClient.UnshareServiceInstanceFromSpace(serviceInstance.GUID, unshareSpace.GUID)
			return
		},
	))
}

func (actor Actor) validateSharingDetails(
	serviceInstanceName, targetedSpaceGUID, targetedOrgGUID string,
	sharedToDetails ServiceInstanceSharingParams,
) (resources.ServiceInstance, resources.Space, ccv3.Warnings, error) {
	var serviceInstance resources.ServiceInstance
	var shareSpace resources.Space
	var shareToOrgGUID = targetedOrgGUID

	warnings, err := railway.Sequentially(
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
	)

	if err != nil {
		return resources.ServiceInstance{}, resources.Space{}, warnings, err
	}

	return serviceInstance, shareSpace, warnings, nil
}
