package v2action

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Organization represents a CLI Organization.
type Organization ccv2.Organization

// OrganizationNotFoundError represents the scenario when the organization
// searched for could not be found.
type OrganizationNotFoundError struct {
	GUID string
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

// GetOrganization returns an Organization based on the provided guid.
func (actor Actor) GetOrganization(guid string) (Organization, Warnings, error) {
	org, warnings, err := actor.CloudControllerClient.GetOrganization(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return Organization{}, Warnings(warnings), OrganizationNotFoundError{GUID: guid}
	}

	return Organization(org), Warnings(warnings), err
}

// GetOrganizationByName returns an Organization based off of the name given.
func (actor Actor) GetOrganizationByName(orgName string) (Organization, Warnings, error) {
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations(ccv2.Query{
		Filter:   ccv2.NameFilter,
		Operator: ccv2.EqualOperator,
		Values:   []string{orgName},
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

	ccWarnings, err := actor.CloudControllerClient.PollJob(job)
	for _, warning := range ccWarnings {
		allWarnings = append(allWarnings, warning)
	}

	return allWarnings, err
}

func (actor Actor) GetOrganizations() ([]Organization, Warnings, error) {
	var returnedOrgs []Organization
	orgs, warnings, err := actor.CloudControllerClient.GetOrganizations()
	for _, org := range orgs {
		returnedOrgs = append(returnedOrgs, Organization(org))
	}
	return returnedOrgs, Warnings(warnings), err
}
