package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Space ccv3.Space

func (actor Actor) CreateSpace(spaceName, orgGUID string) (Space, Warnings, error) {
	allWarnings := Warnings{}

	space, apiWarnings, err := actor.CloudControllerClient.CreateSpace(ccv3.Space{
		Name: spaceName,
		Relationships: ccv3.Relationships{
			constant.RelationshipTypeOrganization: ccv3.Relationship{GUID: orgGUID},
		},
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if _, ok := err.(ccerror.NameNotUniqueInOrgError); ok {
		return Space{}, allWarnings, actionerror.SpaceAlreadyExistsError{Space: spaceName}
	}
	return Space{
		GUID: space.GUID,
		Name: spaceName,
	}, allWarnings, err
}

// ResetSpaceIsolationSegment disassociates a space from an isolation segment.
//
// If the space's organization has a default isolation segment, return its
// name. Otherwise return the empty string.
func (actor Actor) ResetSpaceIsolationSegment(orgGUID string, spaceGUID string) (string, Warnings, error) {
	var allWarnings Warnings

	_, apiWarnings, err := actor.CloudControllerClient.UpdateSpaceIsolationSegmentRelationship(spaceGUID, "")
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
	ccv3Spaces, warnings, err := actor.CloudControllerClient.GetSpaces(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{spaceName}},
		ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
	)

	if err != nil {
		return Space{}, Warnings(warnings), err
	}

	if len(ccv3Spaces) == 0 {
		return Space{}, Warnings(warnings), actionerror.SpaceNotFoundError{Name: spaceName}
	}

	return Space(ccv3Spaces[0]), Warnings(warnings), nil
}

// GetOrganizationSpaces returns a list of spaces in the specified org
func (actor Actor) GetOrganizationSpaces(orgGUID string) ([]Space, Warnings, error) {
	ccv3Spaces, warnings, err := actor.CloudControllerClient.GetSpaces(ccv3.Query{
		Key:    ccv3.OrganizationGUIDFilter,
		Values: []string{orgGUID},
	})
	if err != nil {
		return []Space{}, Warnings(warnings), err
	}

	spaces := make([]Space, len(ccv3Spaces))
	for i, ccv3Space := range ccv3Spaces {
		spaces[i] = Space(ccv3Space)
	}

	return spaces, Warnings(warnings), nil
}

func (actor Actor) DeleteSpaceByNameAndOrganizationName(spaceName string, orgName string) (Warnings, error) {
	var allWarnings Warnings

	org, actorWarnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, actorWarnings...)
	if err != nil {
		return allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	jobURL, deleteWarnings, err := actor.CloudControllerClient.DeleteSpace(space.GUID)
	allWarnings = append(allWarnings, Warnings(deleteWarnings)...)
	if err != nil {
		return allWarnings, err
	}

	ccWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, Warnings(ccWarnings)...)

	return allWarnings, err
}
