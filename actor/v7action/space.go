package v7action

import (
	"fmt"
	"sort"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
)

type Space ccv3.Space

type SpaceSummary struct {
	Space
	Name                  string
	OrgName               string
	AppNames              []string
	ServiceInstanceNames  []string
	IsolationSegmentName  string
	QuotaName             string
	RunningSecurityGroups []resources.SecurityGroup
	StagingSecurityGroups []resources.SecurityGroup
}

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
	ccv3Spaces, _, warnings, err := actor.CloudControllerClient.GetSpaces(
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

func (actor Actor) GetSpaceSummaryByNameAndOrganization(spaceName string, orgGUID string) (SpaceSummary, Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByGUID(orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	apps, warnings, err := actor.GetApplicationsBySpace(space.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	appNames := make([]string, len(apps))
	for i, app := range apps {
		appNames[i] = app.Name
	}
	sort.Strings(appNames)

	serviceInstances, ccv3Warnings, err := actor.CloudControllerClient.GetServiceInstances(
		ccv3.Query{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{space.GUID},
		})
	allWarnings = append(allWarnings, Warnings(ccv3Warnings)...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	serviceInstanceNames := make([]string, len(serviceInstances))
	for i, instance := range serviceInstances {
		serviceInstanceNames[i] = instance.Name
	}
	sort.Strings(serviceInstanceNames)

	isoSegRelationship, ccv3Warnings, err := actor.CloudControllerClient.GetSpaceIsolationSegment(space.GUID)
	allWarnings = append(allWarnings, Warnings(ccv3Warnings)...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	isoSegName := ""
	isoSegGUID := isoSegRelationship.GUID
	isDefaultIsoSeg := false

	if isoSegGUID == "" {
		defaultIsoSeg, ccv3Warnings, err := actor.CloudControllerClient.GetOrganizationDefaultIsolationSegment(org.GUID)
		allWarnings = append(allWarnings, Warnings(ccv3Warnings)...)
		if err != nil {
			return SpaceSummary{}, allWarnings, err
		}
		isoSegGUID = defaultIsoSeg.GUID
		if isoSegGUID != "" {
			isDefaultIsoSeg = true
		}
	}

	if isoSegGUID != "" {
		isoSeg, ccv3warnings, err := actor.CloudControllerClient.GetIsolationSegment(isoSegGUID)
		allWarnings = append(allWarnings, Warnings(ccv3warnings)...)
		if err != nil {
			return SpaceSummary{}, allWarnings, err
		}
		if isDefaultIsoSeg {
			isoSegName = fmt.Sprintf("%s (org default)", isoSeg.Name)
		} else {
			isoSegName = isoSeg.Name
		}
	}

	appliedQuotaRelationshipGUID := space.Relationships[constant.RelationshipTypeQuota].GUID

	var ccv3SpaceQuota ccv3.SpaceQuota
	if appliedQuotaRelationshipGUID != "" {
		ccv3SpaceQuota, ccv3Warnings, err = actor.CloudControllerClient.GetSpaceQuota(space.Relationships[constant.RelationshipTypeQuota].GUID)
		allWarnings = append(allWarnings, Warnings(ccv3Warnings)...)

		if err != nil {
			return SpaceSummary{}, allWarnings, err
		}
	}

	runningSecurityGroups, ccv3Warnings, err := actor.CloudControllerClient.GetRunningSecurityGroups(space.GUID)
	allWarnings = append(allWarnings, ccv3Warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	stagingSecurityGroups, ccv3Warnings, err := actor.CloudControllerClient.GetStagingSecurityGroups(space.GUID)
	allWarnings = append(allWarnings, ccv3Warnings...)
	if err != nil {
		return SpaceSummary{}, allWarnings, err
	}

	spaceSummary := SpaceSummary{
		OrgName:               org.Name,
		Name:                  space.Name,
		Space:                 space,
		AppNames:              appNames,
		ServiceInstanceNames:  serviceInstanceNames,
		IsolationSegmentName:  isoSegName,
		QuotaName:             ccv3SpaceQuota.Name,
		RunningSecurityGroups: runningSecurityGroups,
		StagingSecurityGroups: stagingSecurityGroups,
	}

	return spaceSummary, allWarnings, nil
}

// GetOrganizationSpacesWithLabelSelector returns a list of spaces in the specified org
func (actor Actor) GetOrganizationSpacesWithLabelSelector(orgGUID string, labelSelector string) ([]Space, Warnings, error) {

	queries := []ccv3.Query{
		ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.NameOrder}},
	}
	if len(labelSelector) > 0 {
		queries = append(queries, ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}})
	}

	ccv3Spaces, _, warnings, err := actor.CloudControllerClient.GetSpaces(queries...)
	if err != nil {
		return []Space{}, Warnings(warnings), err
	}

	spaces := make([]Space, len(ccv3Spaces))
	for i, ccv3Space := range ccv3Spaces {
		spaces[i] = Space(ccv3Space)
	}

	return spaces, Warnings(warnings), nil
}

// GetOrganizationSpaces returns a list of spaces in the specified org
func (actor Actor) GetOrganizationSpaces(orgGUID string) ([]Space, Warnings, error) {
	return actor.GetOrganizationSpacesWithLabelSelector(orgGUID, "")
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

func (actor Actor) RenameSpaceByNameAndOrganizationGUID(oldSpaceName, newSpaceName, orgGUID string) (Space, Warnings, error) {
	var allWarnings Warnings

	space, getWarnings, err := actor.GetSpaceByNameAndOrganization(oldSpaceName, orgGUID)
	allWarnings = append(allWarnings, getWarnings...)
	if err != nil {
		return Space{}, allWarnings, err
	}

	ccSpace, updateWarnings, err := actor.CloudControllerClient.UpdateSpace(ccv3.Space{GUID: space.GUID, Name: newSpaceName})
	allWarnings = append(allWarnings, Warnings(updateWarnings)...)

	return Space(ccSpace), allWarnings, err
}
