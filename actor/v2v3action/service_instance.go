package v2v3action

import (
	"strings"

	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/actor/v2action"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv2/constant"
)

func (actor Actor) ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganizationName(shareToSpaceName string, serviceInstanceName string, sourceSpaceGUID string, shareToOrgName string) (Warnings, error) {
	var allWarnings Warnings

	organization, warningsV3, err := actor.V3Actor.GetOrganizationByName(shareToOrgName)
	allWarnings = Warnings(warningsV3)
	if err != nil {
		return allWarnings, err
	}

	warnings, err := actor.ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganization(shareToSpaceName, serviceInstanceName, sourceSpaceGUID, organization.GUID)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}

func (actor Actor) ShareServiceInstanceToSpaceNameByNameAndSpaceAndOrganization(shareToSpaceName string, serviceInstanceName string, sourceSpaceGUID string, shareToOrgGUID string) (Warnings, error) {
	var allWarnings Warnings

	serviceInstance, warningsV2, err := actor.V2Actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, sourceSpaceGUID)
	allWarnings = append(allWarnings, warningsV2...)
	if err != nil {
		if _, ok := err.(actionerror.ServiceInstanceNotFoundError); ok {
			return allWarnings, actionerror.SharedServiceInstanceNotFoundError{}
		}
		return allWarnings, err
	}

	shareToSpace, warningsV2, err := actor.V2Actor.GetSpaceByOrganizationAndName(shareToOrgGUID, shareToSpaceName)
	allWarnings = append(allWarnings, warningsV2...)
	if err != nil {
		return allWarnings, err
	}

	if serviceInstance.IsManaged() {
		var warnings Warnings
		_, warnings, err = actor.isServiceInstanceShareableByService(serviceInstance.ServiceGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return allWarnings, err
		}

		var serviceInstanceSharedTos []v2action.ServiceInstanceSharedTo
		serviceInstanceSharedTos, warningsV2, err = actor.V2Actor.GetServiceInstanceSharedTosByServiceInstance(serviceInstance.GUID)
		allWarnings = append(allWarnings, warningsV2...)
		if err != nil {
			return allWarnings, err
		}

		for _, sharedTo := range serviceInstanceSharedTos {
			if sharedTo.SpaceGUID == shareToSpace.GUID {
				return allWarnings, actionerror.ServiceInstanceAlreadySharedError{}
			}
		}
	}

	_, warningsV3, err := actor.V3Actor.ShareServiceInstanceToSpaces(serviceInstance.GUID, []string{shareToSpace.GUID})
	allWarnings = append(allWarnings, warningsV3...)
	return allWarnings, err
}

func (actor Actor) isServiceInstanceShareableByService(serviceGUID string) (bool, Warnings, error) {
	var allWarnings Warnings

	service, warningsV2, err := actor.V2Actor.GetService(serviceGUID)
	allWarnings = append(allWarnings, warningsV2...)
	if err != nil {
		return false, allWarnings, err
	}

	featureFlags, warningsV2, err := actor.V2Actor.GetFeatureFlags()
	allWarnings = append(allWarnings, warningsV2...)
	if err != nil {
		return false, allWarnings, err
	}

	var featureFlagEnabled bool
	for _, flag := range featureFlags {
		if flag.Name == string(constant.FeatureFlagServiceInstanceSharing) {
			featureFlagEnabled = flag.Enabled
		}
	}

	if !featureFlagEnabled || !service.Extra.Shareable {
		return false, allWarnings, actionerror.ServiceInstanceNotShareableError{
			FeatureFlagEnabled:          featureFlagEnabled,
			ServiceBrokerSharingEnabled: service.Extra.Shareable,
		}
	}

	return true, allWarnings, nil
}

func (actor Actor) UnshareServiceInstanceFromOrganizationNameAndSpaceNameByNameAndSpace(sharedToOrgName string, sharedToSpaceName string, serviceInstanceName string, currentlyTargetedSpaceGUID string) (Warnings, error) {
	var allWarnings Warnings
	serviceInstance, warningsV2, err := actor.V2Actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, currentlyTargetedSpaceGUID)
	allWarnings = append(allWarnings, warningsV2...)
	if err != nil {
		if _, ok := err.(actionerror.ServiceInstanceNotFoundError); ok {
			return allWarnings, actionerror.SharedServiceInstanceNotFoundError{}
		}
		return allWarnings, err
	}

	sharedTos, warningsV2, err := actor.V2Actor.GetServiceInstanceSharedTosByServiceInstance(serviceInstance.GUID)
	allWarnings = append(allWarnings, warningsV2...)
	if err != nil {
		return allWarnings, err
	}

	sharedToSpaceGUID := ""
	for _, sharedTo := range sharedTos {
		if strings.EqualFold(sharedTo.SpaceName, sharedToSpaceName) && strings.EqualFold(sharedTo.OrganizationName, sharedToOrgName) {
			sharedToSpaceGUID = sharedTo.SpaceGUID
		}
	}
	if sharedToSpaceGUID == "" {
		return allWarnings, actionerror.ServiceInstanceNotSharedToSpaceError{ServiceInstanceName: serviceInstanceName}
	}

	warningsV3, err := actor.V3Actor.UnshareServiceInstanceByServiceInstanceAndSpace(serviceInstance.GUID, sharedToSpaceGUID)
	allWarnings = append(allWarnings, warningsV3...)
	return allWarnings, err
}
