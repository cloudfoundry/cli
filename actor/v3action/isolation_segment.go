package v3action

import (
	"fmt"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type IsolationSegmentSummary struct {
	Name         string
	EntitledOrgs []string
}

// IsolationSegment represents a V3 actor IsolationSegment.
type IsolationSegment ccv3.IsolationSegment

// IsolationSegmentNotFoundError represents the error that occurs when the
// isolation segment is not found.
type IsolationSegmentNotFoundError struct {
	Name string
}

func (e IsolationSegmentNotFoundError) Error() string {
	return fmt.Sprintf("Isolation Segment '%s' not found.", e.Name)
}

// IsolationSegmentAlreadyExistsError gets returned when an isolation segment
// already exists.
type IsolationSegmentAlreadyExistsError struct {
	Name string
}

func (e IsolationSegmentAlreadyExistsError) Error() string {
	return fmt.Sprintf("Isolation Segment '%s' already exists.", e.Name)
}

// CreateIsolationSegmentByName creates a given isolation segment.
func (actor Actor) CreateIsolationSegmentByName(name string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.CreateIsolationSegment(name)
	if _, ok := err.(cloudcontroller.UnprocessableEntityError); ok {
		return Warnings(warnings), IsolationSegmentAlreadyExistsError{Name: name}
	}
	return Warnings(warnings), err
}

// DeleteIsolationSegmentByName deletes the given isolation segment.
func (actor Actor) DeleteIsolationSegmentByName(name string) (Warnings, error) {
	isolationSegment, warnings, err := actor.GetIsolationSegmentByName(name)
	allWarnings := append(Warnings{}, warnings...)
	if err != nil {
		return allWarnings, err
	}

	apiWarnings, err := actor.CloudControllerClient.DeleteIsolationSegment(isolationSegment.GUID)
	return append(allWarnings, apiWarnings...), err
}

// EntitleIsolationSegmentToOrganizationByName entitles the given organization
// to use the specified isolation segment
func (actor Actor) EntitleIsolationSegmentToOrganizationByName(isolationSegmentName string, orgName string) (Warnings, error) {
	isolationSegment, warnings, err := actor.GetIsolationSegmentByName(isolationSegmentName)
	allWarnings := append(Warnings{}, warnings...)
	if err != nil {
		return allWarnings, err
	}

	organization, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	_, apiWarnings, err := actor.CloudControllerClient.EntitleIsolationSegmentToOrganizations(isolationSegment.GUID, []string{organization.GUID})
	return append(allWarnings, apiWarnings...), err
}

func (actor Actor) AssignIsolationSegmentToSpaceByNameAndSpace(isolationSegmentName string, spaceGUID string) (Warnings, error) {
	seg, warnings, err := actor.GetIsolationSegmentByName(isolationSegmentName)
	if err != nil {
		return warnings, err
	}

	_, apiWarnings, err := actor.CloudControllerClient.AssignSpaceToIsolationSegment(spaceGUID, seg.GUID)
	return append(warnings, apiWarnings...), err
}

// GetIsolationSegmentByName returns the requested isolation segment.
func (actor Actor) GetIsolationSegmentByName(name string) (IsolationSegment, Warnings, error) {
	isolationSegments, warnings, err := actor.CloudControllerClient.GetIsolationSegments(url.Values{ccv3.NameFilter: []string{name}})
	if err != nil {
		return IsolationSegment{}, Warnings(warnings), err
	}

	if len(isolationSegments) == 0 {
		return IsolationSegment{}, Warnings(warnings), IsolationSegmentNotFoundError{Name: name}
	}

	return IsolationSegment(isolationSegments[0]), Warnings(warnings), nil
}

// GetIsolationSegmentSummaries returns all isolation segments and their entitled orgs
func (actor Actor) GetIsolationSegmentSummaries() ([]IsolationSegmentSummary, Warnings, error) {
	isolationSegments, warnings, err := actor.CloudControllerClient.GetIsolationSegments(nil)
	allWarnings := append(Warnings{}, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var isolationSegmentSummaries []IsolationSegmentSummary

	for _, isolationSegment := range isolationSegments {
		isolationSegmentSummary := IsolationSegmentSummary{
			Name:         isolationSegment.Name,
			EntitledOrgs: []string{},
		}

		orgs, warnings, err := actor.CloudControllerClient.GetIsolationSegmentOrganizationsByIsolationSegment(isolationSegment.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}

		for _, org := range orgs {
			isolationSegmentSummary.EntitledOrgs = append(isolationSegmentSummary.EntitledOrgs, org.Name)
		}

		isolationSegmentSummaries = append(isolationSegmentSummaries, isolationSegmentSummary)
	}
	return isolationSegmentSummaries, allWarnings, nil
}

func (actor Actor) GetIsolationSegmentsByOrganization(orgGUID string) ([]IsolationSegment, Warnings, error) {
	ccv3IsolationSegments, warnings, err := actor.CloudControllerClient.GetIsolationSegments(url.Values{
		ccv3.OrganizationGUIDFilter: []string{orgGUID},
	})
	if err != nil {
		return []IsolationSegment{}, Warnings(warnings), err
	}

	isolationSegments := make([]IsolationSegment, len(ccv3IsolationSegments))

	for i, _ := range ccv3IsolationSegments {
		isolationSegments[i] = IsolationSegment(ccv3IsolationSegments[i])
	}

	return isolationSegments, Warnings(warnings), nil
}

func (actor Actor) RevokeIsolationSegmentFromOrganizationByName(isolationSegmentName string, orgName string) (Warnings, error) {
	segment, warnings, err := actor.GetIsolationSegmentByName(isolationSegmentName)
	allWarnings := append(Warnings{}, warnings...)
	if err != nil {
		return allWarnings, err
	}

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}

	apiWarnings, err := actor.CloudControllerClient.RevokeIsolationSegmentFromOrganization(segment.GUID, org.GUID)

	allWarnings = append(allWarnings, apiWarnings...)
	return allWarnings, err
}
