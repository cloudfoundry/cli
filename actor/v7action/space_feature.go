package v7action

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
