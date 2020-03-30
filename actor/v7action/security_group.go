package v7action

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
)

type SecurityGroupSummary struct {
	Name                string
	Rules               []resources.Rule
	SecurityGroupSpaces []SecurityGroupSpace
}

type SecurityGroupSpace struct {
	SpaceName string
	OrgName   string
	Lifecycle string
}

func (actor Actor) BindSecurityGroupToSpace(securityGroupGUID string, spaceGUID string, lifecycle constant.SecurityGroupLifecycle) (Warnings, error) {
	var (
		warnings ccv3.Warnings
		err      error
	)

	switch lifecycle {
	case constant.SecurityGroupLifecycleRunning:
		warnings, err = actor.CloudControllerClient.UpdateSecurityGroupRunningSpace(securityGroupGUID, spaceGUID)
	case constant.SecurityGroupLifecycleStaging:
		warnings, err = actor.CloudControllerClient.UpdateSecurityGroupStagingSpace(securityGroupGUID, spaceGUID)
	default:
		err = actionerror.InvalidLifecycleError{Lifecycle: string(lifecycle)}
	}

	return Warnings(warnings), err
}

func (actor Actor) CreateSecurityGroup(name, filePath string) (Warnings, error) {
	allWarnings := Warnings{}
	bytes, err := parsePath(filePath)
	if err != nil {
		return allWarnings, err
	}

	var rules []resources.Rule
	err = json.Unmarshal(bytes, &rules)
	if err != nil {
		return allWarnings, err
	}

	securityGroup := resources.SecurityGroup{Name: name, Rules: rules}

	_, warnings, err := actor.CloudControllerClient.CreateSecurityGroup(securityGroup)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}

func (actor Actor) GetSecurityGroup(securityGroupName string) (resources.SecurityGroup, Warnings, error) {
	allWarnings := Warnings{}

	securityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups(ccv3.Query{Key: ccv3.NameFilter, Values: []string{securityGroupName}})
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return resources.SecurityGroup{}, allWarnings, err
	}

	if len(securityGroups) == 0 {
		return resources.SecurityGroup{}, allWarnings, actionerror.SecurityGroupNotFoundError{Name: securityGroupName}
	}

	return securityGroups[0], allWarnings, err
}

func (actor Actor) GetSecurityGroupSummary(securityGroupName string) (SecurityGroupSummary, Warnings, error) {
	allWarnings := Warnings{}
	securityGroupSummary := SecurityGroupSummary{}
	securityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups(ccv3.Query{Key: ccv3.NameFilter, Values: []string{securityGroupName}})

	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return securityGroupSummary, allWarnings, err
	}
	if len(securityGroups) == 0 {
		return securityGroupSummary, allWarnings, actionerror.SecurityGroupNotFoundError{Name: securityGroupName}
	}

	securityGroupSummary.Name = securityGroupName
	securityGroupSummary.Rules = securityGroups[0].Rules

	var noSecurityGroupSpaces = len(securityGroups[0].StagingSpaceGUIDs) == 0 && len(securityGroups[0].RunningSpaceGUIDs) == 0
	if noSecurityGroupSpaces {
		securityGroupSummary.SecurityGroupSpaces = []SecurityGroupSpace{}
	} else {
		secGroupSpaces, warnings, err := getSecurityGroupSpaces(actor, securityGroups[0].StagingSpaceGUIDs, securityGroups[0].RunningSpaceGUIDs)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return securityGroupSummary, allWarnings, err
		}
		securityGroupSummary.SecurityGroupSpaces = secGroupSpaces
	}

	return securityGroupSummary, allWarnings, nil
}

func (actor Actor) GetSecurityGroups() ([]SecurityGroupSummary, Warnings, error) {
	allWarnings := Warnings{}
	securityGroupSummaries := []SecurityGroupSummary{}
	securityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups()

	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return securityGroupSummaries, allWarnings, err
	}

	for _, securityGroup := range securityGroups {
		var securityGroupSummary SecurityGroupSummary
		securityGroupSummary.Name = securityGroup.Name
		securityGroupSummary.Rules = securityGroup.Rules

		var securityGroupSpaces []SecurityGroupSpace
		var noSecurityGroupSpaces = !*securityGroup.StagingGloballyEnabled && !*securityGroup.RunningGloballyEnabled && len(securityGroup.StagingSpaceGUIDs) == 0 && len(securityGroup.RunningSpaceGUIDs) == 0
		if noSecurityGroupSpaces {
			securityGroupSpaces = []SecurityGroupSpace{}
		}

		if *securityGroup.StagingGloballyEnabled {
			securityGroupSpaces = append(securityGroupSpaces, SecurityGroupSpace{
				SpaceName: "<all>",
				OrgName:   "<all>",
				Lifecycle: "staging",
			})
		}

		if *securityGroup.RunningGloballyEnabled {
			securityGroupSpaces = append(securityGroupSpaces, SecurityGroupSpace{
				SpaceName: "<all>",
				OrgName:   "<all>",
				Lifecycle: "running",
			})
		}

		secGroupSpaces, warnings, err := getSecurityGroupSpaces(actor, securityGroup.StagingSpaceGUIDs, securityGroup.RunningSpaceGUIDs)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return securityGroupSummaries, allWarnings, err
		}
		securityGroupSpaces = append(securityGroupSpaces, secGroupSpaces...)
		securityGroupSummary.SecurityGroupSpaces = securityGroupSpaces

		securityGroupSummaries = append(securityGroupSummaries, securityGroupSummary)
	}

	return securityGroupSummaries, allWarnings, nil
}

func (actor Actor) UnbindSecurityGroup(securityGroupName string, orgName string, spaceName string, lifecycle constant.SecurityGroupLifecycle) (Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, org.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	securityGroup, warnings, err := actor.GetSecurityGroup(securityGroupName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	var ccv3Warnings ccv3.Warnings
	switch lifecycle {
	case constant.SecurityGroupLifecycleRunning:
		ccv3Warnings, err = actor.CloudControllerClient.UnbindSecurityGroupRunningSpace(securityGroup.GUID, space.GUID)
	case constant.SecurityGroupLifecycleStaging:
		ccv3Warnings, err = actor.CloudControllerClient.UnbindSecurityGroupStagingSpace(securityGroup.GUID, space.GUID)
	default:
		return allWarnings, actionerror.InvalidLifecycleError{Lifecycle: string(lifecycle)}
	}

	allWarnings = append(allWarnings, ccv3Warnings...)

	if err != nil {
		if _, isNotBoundError := err.(ccerror.SecurityGroupNotBound); isNotBoundError {
			return allWarnings, actionerror.SecurityGroupNotBoundToSpaceError{
				Name:      securityGroupName,
				Space:     spaceName,
				Lifecycle: lifecycle,
			}
		}
	}

	return allWarnings, err
}

func (actor Actor) GetGlobalStagingSecurityGroups() ([]resources.SecurityGroup, Warnings, error) {
	stagingSecurityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups(ccv3.Query{Key: ccv3.GloballyEnabledStaging, Values: []string{"true"}})

	return stagingSecurityGroups, Warnings(warnings), err
}

func (actor Actor) GetGlobalRunningSecurityGroups() ([]resources.SecurityGroup, Warnings, error) {
	runningSecurityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups(ccv3.Query{Key: ccv3.GloballyEnabledRunning, Values: []string{"true"}})

	return runningSecurityGroups, Warnings(warnings), err
}

func (actor Actor) UpdateSecurityGroup(name, filePath string) (Warnings, error) {
	allWarnings := Warnings{}

	// parse input file
	bytes, err := parsePath(filePath)
	if err != nil {
		return allWarnings, err
	}

	var rules []resources.Rule
	err = json.Unmarshal(bytes, &rules)
	if err != nil {
		return allWarnings, err
	}

	// fetch security group from API
	securityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups(ccv3.Query{Key: ccv3.NameFilter, Values: []string{name}})
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	if len(securityGroups) == 0 {
		return allWarnings, actionerror.SecurityGroupNotFoundError{Name: name}
	}

	securityGroup := resources.SecurityGroup{
		Name:  name,
		GUID:  securityGroups[0].GUID,
		Rules: rules,
	}

	// update security group
	_, warnings, err = actor.CloudControllerClient.UpdateSecurityGroup(securityGroup)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}

func (actor Actor) UpdateSecurityGroupGloballyEnabled(securityGroupName string, lifecycle constant.SecurityGroupLifecycle, enabled bool) (Warnings, error) {
	var allWarnings Warnings

	securityGroup, warnings, err := actor.GetSecurityGroup(securityGroupName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	requestBody := resources.SecurityGroup{GUID: securityGroup.GUID}
	switch lifecycle {
	case constant.SecurityGroupLifecycleRunning:
		requestBody.RunningGloballyEnabled = &enabled
	case constant.SecurityGroupLifecycleStaging:
		requestBody.StagingGloballyEnabled = &enabled
	default:
		return allWarnings, actionerror.InvalidLifecycleError{Lifecycle: string(lifecycle)}
	}

	_, ccv3Warnings, err := actor.CloudControllerClient.UpdateSecurityGroup(requestBody)
	allWarnings = append(allWarnings, ccv3Warnings...)

	return allWarnings, err
}

func (actor Actor) DeleteSecurityGroup(securityGroupName string) (Warnings, error) {
	var allWarnings Warnings

	securityGroup, warnings, err := actor.GetSecurityGroup(securityGroupName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	jobURL, ccv3Warnings, err := actor.CloudControllerClient.DeleteSecurityGroup(securityGroup.GUID)
	allWarnings = append(allWarnings, ccv3Warnings...)
	if err != nil {
		return allWarnings, err
	}

	pollingWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, pollingWarnings...)
	return allWarnings, err
}

func getSecurityGroupSpaces(actor Actor, stagingSpaceGUIDs []string, runningSpaceGUIDs []string) ([]SecurityGroupSpace, ccv3.Warnings, error) {
	var warnings ccv3.Warnings
	associatedSpaceGuids := runningSpaceGUIDs
	associatedSpaceGuids = append(associatedSpaceGuids, stagingSpaceGUIDs...)

	var securityGroupSpaces []SecurityGroupSpace
	if len(associatedSpaceGuids) > 0 {
		spaces, includes, newWarnings, err := actor.CloudControllerClient.GetSpaces(ccv3.Query{
			Key:    ccv3.GUIDFilter,
			Values: associatedSpaceGuids,
		}, ccv3.Query{
			Key:    ccv3.Include,
			Values: []string{"organization"},
		})

		warnings = newWarnings
		if err != nil {
			return securityGroupSpaces, warnings, err
		}

		orgsByGuid := make(map[string]ccv3.Organization)
		for _, org := range includes.Organizations {
			orgsByGuid[org.GUID] = org
		}

		spacesByGuid := make(map[string]ccv3.Space)
		for _, space := range spaces {
			spacesByGuid[space.GUID] = space
		}

		for _, runningSpaceGUID := range runningSpaceGUIDs {
			space := spacesByGuid[runningSpaceGUID]
			orgGuid := space.Relationships[constant.RelationshipTypeOrganization].GUID
			securityGroupSpaces = append(securityGroupSpaces, SecurityGroupSpace{
				SpaceName: space.Name,
				OrgName:   orgsByGuid[orgGuid].Name,
				Lifecycle: "running",
			})
		}

		for _, stagingSpaceGUID := range stagingSpaceGUIDs {
			space := spacesByGuid[stagingSpaceGUID]
			orgGuid := space.Relationships[constant.RelationshipTypeOrganization].GUID
			securityGroupSpaces = append(securityGroupSpaces, SecurityGroupSpace{
				SpaceName: space.Name,
				OrgName:   orgsByGuid[orgGuid].Name,
				Lifecycle: "staging",
			})
		}
	}
	return securityGroupSpaces, warnings, nil
}

func parsePath(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
