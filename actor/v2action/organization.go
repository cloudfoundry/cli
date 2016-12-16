package v2action

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Organization represents a CLI Organization.
type Organization ccv2.Organization

// OrganizationNotFoundError represents the scenario when the organization
// searched for could not be found.
type OrganizationNotFoundError struct {
	Name string
}

func (e OrganizationNotFoundError) Error() string {
	return fmt.Sprintf("Organization '%s' not found.", e.Name)
}

// MultipleOrganizationsFoundError represents the scenario when the cloud
// controller returns multiple organizations when filtering by name. This is a
// far out edge case and should not happen.
type MultipleOrganizationsFoundError struct {
	Name  string
	GUIDs []string
}

func (e MultipleOrganizationsFoundError) Error() string {
	guids := strings.Join(e.GUIDs, ", ")
	return fmt.Sprintf("Organization name '%s' matches multiple GUIDs: %s", e.Name, guids)
}

// DeleteOrganization deletes the Organization associated with the provided
// GUID.
func (actor Actor) DeleteOrganization(orgName string) (Warnings, error) {
	orgs, getWarnings, err := actor.CloudControllerClient.GetOrganizations([]ccv2.Query{
		{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.EqualOperator,
			Value:    orgName,
		},
	})
	if err != nil {
		return Warnings(getWarnings), err
	}

	if len(orgs) == 0 {
		return Warnings(getWarnings), OrganizationNotFoundError{Name: orgName}
	}

	if len(orgs) > 1 {
		var guids []string
		for _, org := range orgs {
			guids = append(guids, org.GUID)
		}
		return Warnings(getWarnings), MultipleOrganizationsFoundError{Name: orgName, GUIDs: guids}
	}

	var allWarnings Warnings
	allWarnings = append(allWarnings, getWarnings...)

	deleteWarnings, err := actor.CloudControllerClient.DeleteOrganization(orgs[0].GUID)
	allWarnings = append(allWarnings, deleteWarnings...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}
