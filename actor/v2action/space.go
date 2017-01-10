package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Space represents a CLI Space
type Space ccv2.Space

// SpaceFoundError represents the scenario when the space searched for could
// not be found.
type SpaceNotFoundError struct {
	Name string
}

func (e SpaceNotFoundError) Error() string {
	return fmt.Sprintf("Space '%s' not found.", e.Name)
}

// MultipleSpacesFoundError represents the scenario when the cloud
// controller returns multiple spaces when filtering by name. This is a
// far out edge case and should not happen.
type MultipleSpacesFoundError struct {
	SpaceName string
	OrgGUID   string
}

func (e MultipleSpacesFoundError) Error() string {
	return fmt.Sprintf("Multiple spaces found matching organization GUID '%s' and name '%s'", e.OrgGUID, e.SpaceName)
}

// GetSpaceByOrganizationAndName returns an Space based on the org and name.
func (actor Actor) GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (Space, Warnings, error) {
	query := []ccv2.Query{
		{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.EqualOperator,
			Value:    spaceName,
		},
		{
			Filter:   ccv2.OrganizationGUIDFilter,
			Operator: ccv2.EqualOperator,
			Value:    orgGUID,
		},
	}

	ccv2Spaces, warnings, err := actor.CloudControllerClient.GetSpaces(query)
	if err != nil {
		return Space{}, Warnings(warnings), err
	}

	if len(ccv2Spaces) == 0 {
		return Space{}, Warnings(warnings), SpaceNotFoundError{Name: spaceName}
	}

	if len(ccv2Spaces) > 1 {
		return Space{}, Warnings(warnings), MultipleSpacesFoundError{OrgGUID: orgGUID, SpaceName: spaceName}
	}

	return Space{
		GUID:     ccv2Spaces[0].GUID,
		Name:     ccv2Spaces[0].Name,
		AllowSSH: ccv2Spaces[0].AllowSSH,
	}, Warnings(warnings), nil
}

// GetOrganizationSpaces returns a list of spaces in the specified org
func (actor Actor) GetOrganizationSpaces(orgGUID string) ([]Space, Warnings, error) {
	query := []ccv2.Query{
		{
			Filter:   ccv2.OrganizationGUIDFilter,
			Operator: ccv2.EqualOperator,
			Value:    orgGUID,
		}}
	ccv2Spaces, warnings, err := actor.CloudControllerClient.GetSpaces(query)
	if err != nil {
		return []Space{}, Warnings(warnings), err
	}

	var spaces []Space

	for _, ccv2Space := range ccv2Spaces {
		spaces = append(spaces, Space{
			GUID:     ccv2Space.GUID,
			Name:     ccv2Space.Name,
			AllowSSH: ccv2Space.AllowSSH,
		})
	}

	return spaces, Warnings(warnings), nil
}
