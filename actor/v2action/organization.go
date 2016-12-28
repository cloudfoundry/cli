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

// GetOrganizationByName returns an Organization based off of the name given.
func (actor Actor) GetOrganizationByName(orgName string) (Organization, Warnings, error) {
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations([]ccv2.Query{
		{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.EqualOperator,
			Value:    orgName,
		},
	})
	if err != nil {
		return Organization{}, Warnings(warnings), err
	}

	if len(orgs) == 0 {
		return Organization{}, Warnings(warnings), OrganizationNotFoundError{Name: orgName}
	}

	if len(orgs) > 1 {
		var guids []string
		for _, org := range orgs {
			guids = append(guids, org.GUID)
		}
		return Organization{}, Warnings(warnings), MultipleOrganizationsFoundError{Name: orgName, GUIDs: guids}
	}

	return Organization(orgs[0]), Warnings(warnings), nil
}

// DeleteOrganization deletes the Organization associated with the provided
// GUID. Once the deletion request is sent, it polls the deletion job until
// it's finished.
func (actor Actor) DeleteOrganization(orgName string) (Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	job, deleteWarnings, err := actor.CloudControllerClient.DeleteOrganization(org.GUID)
	allWarnings = append(allWarnings, deleteWarnings...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.PollJob(job)
	allWarnings = append(allWarnings, warnings...)

	return allWarnings, err
}
