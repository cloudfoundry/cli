package v3action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Space ccv3.Space

// ResetSpaceIsolationSegment disassociates a space from an isolation segment.
//
// If the space's organization has a default isolation segment, return its
// name. Otherwise return the empty string.
func (actor Actor) ResetSpaceIsolationSegment(orgGUID string, spaceGUID string) (string, Warnings, error) {
	var allWarnings Warnings

	_, apiWarnings, err := actor.CloudControllerClient.AssignSpaceToIsolationSegment(spaceGUID, "")
	allWarnings = append(allWarnings, apiWarnings...)
	if err != nil {
		return "", allWarnings, err
	}

	isoSegRelationship, apiWarnings, err := actor.CloudControllerClient.GetOrganizationDefaultIsolationSegment(orgGUID)
	allWarnings = append(allWarnings, apiWarnings...)
	if err != nil {
		return "", allWarnings, err
	}

	var isoSegName string
	if isoSegRelationship.GUID != "" {
		isolationSegment, apiWarnings, err := actor.CloudControllerClient.GetIsolationSegment(isoSegRelationship.GUID)
		allWarnings = append(allWarnings, apiWarnings...)
		if err != nil {
			return "", allWarnings, err
		}
		isoSegName = isolationSegment.Name
	}

	return isoSegName, allWarnings, nil
}

func (actor Actor) GetSpaceByNameAndOrganization(spaceName string, orgGUID string) (Space, Warnings, error) {
	spaces, warnings, err := actor.CloudControllerClient.GetSpaces(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceName}},
		ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
	)

	if err != nil {
		return Space{}, Warnings(warnings), err
	}

	if len(spaces) == 0 {
		return Space{}, Warnings(warnings), actionerror.SpaceNotFoundError{Name: spaceName}
	}

	return Space(spaces[0]), Warnings(warnings), nil
}
