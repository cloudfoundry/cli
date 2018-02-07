package v2v3action

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
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

	_, warnings, err := actor.isServiceInstanceShareableByService(serviceInstance.ServiceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	serviceInstanceSharedTos, warningsV2, err := actor.V2Actor.GetServiceInstanceSharedTosByServiceInstance(serviceInstance.GUID)
	allWarnings = append(allWarnings, warningsV2...)
	if err != nil {
		return allWarnings, err
	}

	shareToSpace, warningsV2, err := actor.V2Actor.GetSpaceByOrganizationAndName(shareToOrgGUID, shareToSpaceName)
	allWarnings = append(allWarnings, warningsV2...)
	if err != nil {
		return allWarnings, err
	}

	for _, sharedTo := range serviceInstanceSharedTos {
		if sharedTo.SpaceGUID == shareToSpace.GUID {
			allWarnings = append(allWarnings, fmt.Sprintf("Service instance %s is already shared with that space.", serviceInstanceName))
			return allWarnings, actionerror.ServiceInstanceAlreadySharedError{}
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
		if flag.Name == string(ccv2.FeatureFlagServiceInstanceSharing) {
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
