package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
)

type IsolationSegmentSummary struct {
	Name         string
	EntitledOrgs []string
}

// GetEffectiveIsolationSegmentBySpace returns the space's effective isolation
// segment.
//
// If the space has its own isolation segment, that will be returned.
//
// If the space does not have one, the organization's default isolation segment
// (GUID passed in) will be returned.
//
// If the space does not have one and the passed in organization default
// isolation segment GUID is empty, a NoRelationshipError will be returned.
func (actor Actor) GetEffectiveIsolationSegmentBySpace(spaceGUID string, orgDefaultIsolationSegmentGUID string) (resources.IsolationSegment, Warnings, error) {
	relationship, warnings, err := actor.CloudControllerClient.GetSpaceIsolationSegment(spaceGUID)
	allWarnings := append(Warnings{}, warnings...)
	if err != nil {
		return resources.IsolationSegment{}, allWarnings, err
	}

	effectiveGUID := relationship.GUID
	if effectiveGUID == "" {
		if orgDefaultIsolationSegmentGUID != "" {
			effectiveGUID = orgDefaultIsolationSegmentGUID
		} else {
			return resources.IsolationSegment{}, allWarnings, actionerror.NoRelationshipError{}
		}
	}

	isolationSegment, warnings, err := actor.CloudControllerClient.GetIsolationSegment(effectiveGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.IsolationSegment{}, allWarnings, err
	}

	return resources.IsolationSegment(isolationSegment), allWarnings, err
}

// CreateIsolationSegmentByName creates a given isolation segment.
func (actor Actor) CreateIsolationSegmentByName(isolationSegment resources.IsolationSegment) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.CreateIsolationSegment(resources.IsolationSegment(isolationSegment))
	if _, ok := err.(ccerror.UnprocessableEntityError); ok {
		return Warnings(warnings), actionerror.IsolationSegmentAlreadyExistsError{Name: isolationSegment.Name}
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

	_, apiWarnings, err := actor.CloudControllerClient.UpdateSpaceIsolationSegmentRelationship(spaceGUID, seg.GUID)
	return append(warnings, apiWarnings...), err
}

// GetIsolationSegmentByName returns the requested isolation segment.
func (actor Actor) GetIsolationSegmentByName(name string) (resources.IsolationSegment, Warnings, error) {
	isolationSegments, warnings, err := actor.CloudControllerClient.GetIsolationSegments(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{name}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)
	if err != nil {
		return resources.IsolationSegment{}, Warnings(warnings), err
	}

	if len(isolationSegments) == 0 {
		return resources.IsolationSegment{}, Warnings(warnings), actionerror.IsolationSegmentNotFoundError{Name: name}
	}

	return resources.IsolationSegment(isolationSegments[0]), Warnings(warnings), nil
}

// GetIsolationSegmentSummaries returns all isolation segments and their entitled orgs
func (actor Actor) GetIsolationSegmentSummaries() ([]IsolationSegmentSummary, Warnings, error) {
	isolationSegments, warnings, err := actor.CloudControllerClient.GetIsolationSegments()
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

		orgs, warnings, err := actor.CloudControllerClient.GetIsolationSegmentOrganizations(isolationSegment.GUID)
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

func (actor Actor) GetIsolationSegmentsByOrganization(orgGUID string) ([]resources.IsolationSegment, Warnings, error) {
	isolationSegments, warnings, err := actor.CloudControllerClient.GetIsolationSegments(
		ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
	)
	if err != nil {
		return []resources.IsolationSegment{}, Warnings(warnings), err
	}

	return isolationSegments, Warnings(warnings), nil
}

func (actor Actor) DeleteIsolationSegmentOrganizationByName(isolationSegmentName string, orgName string) (Warnings, error) {
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

	apiWarnings, err := actor.CloudControllerClient.DeleteIsolationSegmentOrganization(segment.GUID, org.GUID)

	allWarnings = append(allWarnings, apiWarnings...)
	return allWarnings, err
}

// GetOrganizationDefaultIsolationSegment gets a default isolation segment on
// an organization.
func (actor Actor) GetOrganizationDefaultIsolationSegment(orgGUID string) (string, Warnings, error) {
	defaultIsoSegRelationship, apiWarnings, err := actor.CloudControllerClient.GetOrganizationDefaultIsolationSegment(orgGUID)
	return defaultIsoSegRelationship.GUID, Warnings(apiWarnings), err
}

// SetOrganizationDefaultIsolationSegment sets a default isolation segment on
// an organization.
func (actor Actor) SetOrganizationDefaultIsolationSegment(orgGUID string, isoSegGUID string) (Warnings, error) {
	_, apiWarnings, err := actor.CloudControllerClient.UpdateOrganizationDefaultIsolationSegmentRelationship(orgGUID, isoSegGUID)
	return Warnings(apiWarnings), err
}

// ResetOrganizationDefaultIsolationSegment resets the default isolation segment fon
// an organization.
func (actor Actor) ResetOrganizationDefaultIsolationSegment(orgGUID string) (Warnings, error) {
	_, apiWarnings, err := actor.CloudControllerClient.UpdateOrganizationDefaultIsolationSegmentRelationship(orgGUID, "")
	return Warnings(apiWarnings), err
}
