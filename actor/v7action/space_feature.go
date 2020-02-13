package v7action

import "code.cloudfoundry.org/cli/actor/actionerror"

func (actor Actor) AllowSpaceSSH(spaceName string, orgGUID string) (Warnings, error) {
	var allWarnings Warnings

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	enabled, ccv3Warnings, err := actor.CloudControllerClient.GetSpaceFeature(space.GUID, "ssh")

	allWarnings = append(allWarnings, ccv3Warnings...)

	if err != nil {
		return allWarnings, err
	}

	if enabled {
		return allWarnings, actionerror.SpaceSSHAlreadyEnabledError{Space: space.Name}
	}

	ccv3Warnings, err = actor.CloudControllerClient.UpdateSpaceFeature(space.GUID, true, "ssh")
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
