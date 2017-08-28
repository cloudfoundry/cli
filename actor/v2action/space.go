package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Space represents a CLI Space
type Space ccv2.Space

// SpaceNotFoundError represents the scenario when the space searched for could
// not be found.
type SpaceNotFoundError struct {
	GUID string
	Name string
}

func (e SpaceNotFoundError) Error() string {
	switch {
	case e.Name != "":
		return fmt.Sprintf("Space '%s' not found.", e.Name)
	case e.GUID != "":
		return fmt.Sprintf("Space with GUID '%s' not found.", e.GUID)
	default:
		return fmt.Sprintf("Space '' not found.")
	}
}

// MultipleSpacesFoundError represents the scenario when the cloud
// controller returns multiple spaces when filtering by name. This is a
// far out edge case and should not happen.
type MultipleSpacesFoundError struct {
	Name    string
	OrgGUID string
}

func (e MultipleSpacesFoundError) Error() string {
	return fmt.Sprintf("Multiple spaces found matching organization GUID '%s' and name '%s'", e.OrgGUID, e.Name)
}

func (actor Actor) DeleteSpaceByNameAndOrganizationName(spaceName string, orgName string) (Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByOrganizationAndName(org.GUID, spaceName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	job, deleteWarnings, err := actor.CloudControllerClient.DeleteSpace(space.GUID)
	allWarnings = append(allWarnings, Warnings(deleteWarnings)...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.PollJob(Job(job))
	allWarnings = append(allWarnings, Warnings(warnings)...)

	return allWarnings, err
}

// GetOrganizationSpaces returns a list of spaces in the specified org
func (actor Actor) GetOrganizationSpaces(orgGUID string) ([]Space, Warnings, error) {
	ccv2Spaces, warnings, err := actor.CloudControllerClient.GetSpaces(ccv2.Query{
		Filter:   ccv2.OrganizationGUIDFilter,
		Operator: ccv2.EqualOperator,
		Values:   []string{orgGUID},
	})
	if err != nil {
		return []Space{}, Warnings(warnings), err
	}

	spaces := make([]Space, len(ccv2Spaces))
	for i, ccv2Space := range ccv2Spaces {
		spaces[i] = Space(ccv2Space)
	}

	return spaces, Warnings(warnings), nil
}

// GetSpaceByOrganizationAndName returns an Space based on the org and name.
func (actor Actor) GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (Space, Warnings, error) {
	ccv2Spaces, warnings, err := actor.CloudControllerClient.GetSpaces(
		ccv2.Query{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.EqualOperator,
			Values:   []string{spaceName},
		},
		ccv2.Query{
			Filter:   ccv2.OrganizationGUIDFilter,
			Operator: ccv2.EqualOperator,
			Values:   []string{orgGUID},
		},
	)
	if err != nil {
		return Space{}, Warnings(warnings), err
	}

	if len(ccv2Spaces) == 0 {
		return Space{}, Warnings(warnings), SpaceNotFoundError{Name: spaceName}
	}

	if len(ccv2Spaces) > 1 {
		return Space{}, Warnings(warnings), MultipleSpacesFoundError{OrgGUID: orgGUID, Name: spaceName}
	}

	return Space(ccv2Spaces[0]), Warnings(warnings), nil
}
