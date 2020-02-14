package v7action

import "code.cloudfoundry.org/cli/actor/actionerror"

func (actor Actor) UpdateSpaceFeature(spaceName string, orgGUID string, enabled bool, feature string) (Warnings, error) {
	var allWarnings Warnings

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	previousValue, ccv3Warnings, err := actor.CloudControllerClient.GetSpaceFeature(space.GUID, feature)

	allWarnings = append(allWarnings, ccv3Warnings...)

	if err != nil {
		return allWarnings, err
	}

	if (previousValue == enabled) && (feature == "ssh") {
		if enabled {
			return allWarnings, actionerror.SpaceSSHAlreadyEnabledError{Space: space.Name}
		} else {
			return allWarnings, actionerror.SpaceSSHAlreadyDisabledError{Space: space.Name}
		}
	}

	ccv3Warnings, err = actor.CloudControllerClient.UpdateSpaceFeature(space.GUID, enabled, feature)
	allWarnings = append(allWarnings, Warnings(ccv3Warnings)...)

	return allWarnings, err
}

func (actor Actor) GetSpaceFeature(spaceName string, orgGUID string, feature string) (bool, Warnings, error) {
	var allWarnings Warnings

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return false, allWarnings, err
	}

	enabled, ccv3Warnings, err := actor.CloudControllerClient.GetSpaceFeature(space.GUID, feature)
	allWarnings = append(allWarnings, Warnings(ccv3Warnings)...)

	return enabled, allWarnings, err
}
