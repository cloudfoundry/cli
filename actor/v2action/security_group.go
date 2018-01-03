package v2action

import (
	"sort"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// SecurityGroup represents a CF SecurityGroup.
type SecurityGroup ccv2.SecurityGroup

// SecurityGroupWithOrganizationSpaceAndLifecycle represents a security group with
// organization and space information.
type SecurityGroupWithOrganizationSpaceAndLifecycle struct {
	SecurityGroup *SecurityGroup
	Organization  *Organization
	Space         *Space
	Lifecycle     ccv2.SecurityGroupLifecycle
}

func (actor Actor) BindSecurityGroupToSpace(securityGroupGUID string, spaceGUID string, lifecycle ccv2.SecurityGroupLifecycle) (Warnings, error) {
	var (
		warnings ccv2.Warnings
		err      error
	)

	switch lifecycle {
	case ccv2.SecurityGroupLifecycleRunning:
		warnings, err = actor.CloudControllerClient.AssociateSpaceWithRunningSecurityGroup(securityGroupGUID, spaceGUID)
	case ccv2.SecurityGroupLifecycleStaging:
		warnings, err = actor.CloudControllerClient.AssociateSpaceWithStagingSecurityGroup(securityGroupGUID, spaceGUID)
	default:
		err = actionerror.InvalidLifecycleError{Lifecycle: lifecycle}
	}

	return Warnings(warnings), err
}

func (actor Actor) GetSecurityGroupByName(securityGroupName string) (SecurityGroup, Warnings, error) {
	securityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups(ccv2.QQuery{
		Filter:   ccv2.NameFilter,
		Operator: ccv2.EqualOperator,
		Values:   []string{securityGroupName},
	})

	if err != nil {
		return SecurityGroup{}, Warnings(warnings), err
	}

	if len(securityGroups) == 0 {
		return SecurityGroup{}, Warnings(warnings), actionerror.SecurityGroupNotFoundError{Name: securityGroupName}
	}

	securityGroup := SecurityGroup{
		Name: securityGroups[0].Name,
		GUID: securityGroups[0].GUID,
	}
	return securityGroup, Warnings(warnings), nil
}

type SpaceWithLifecycle struct {
	ccv2.Space
	Lifecycle ccv2.SecurityGroupLifecycle
}

func (actor Actor) getSecurityGroupSpacesAndAssignedLifecycles(securityGroupGUID string, includeStaging bool) ([]SpaceWithLifecycle, Warnings, error) {
	var (
		spacesWithLifecycles []SpaceWithLifecycle
		allWarnings          Warnings
	)

	runningSpaces, warnings, err := actor.CloudControllerClient.GetRunningSpacesBySecurityGroup(securityGroupGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, Warnings(allWarnings), err
	}

	for _, space := range runningSpaces {
		spacesWithLifecycles = append(spacesWithLifecycles, SpaceWithLifecycle{Space: space, Lifecycle: ccv2.SecurityGroupLifecycleRunning})
	}

	if includeStaging {
		stagingSpaces, warnings, err := actor.CloudControllerClient.GetStagingSpacesBySecurityGroup(securityGroupGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, Warnings(allWarnings), err
		}

		for _, space := range stagingSpaces {
			spacesWithLifecycles = append(spacesWithLifecycles, SpaceWithLifecycle{Space: space, Lifecycle: ccv2.SecurityGroupLifecycleStaging})
		}
	}

	return spacesWithLifecycles, allWarnings, nil
}

// GetSecurityGroupsWithOrganizationSpaceAndLifecycle returns a list of security groups
// with org and space information, optionally including staging spaces.
func (actor Actor) GetSecurityGroupsWithOrganizationSpaceAndLifecycle(includeStaging bool) ([]SecurityGroupWithOrganizationSpaceAndLifecycle, Warnings, error) {
	securityGroups, allWarnings, err := actor.CloudControllerClient.GetSecurityGroups()
	if err != nil {
		return nil, Warnings(allWarnings), err
	}

	cachedOrgs := make(map[string]Organization)
	var secGroupOrgSpaces []SecurityGroupWithOrganizationSpaceAndLifecycle

	for _, s := range securityGroups {
		securityGroup := SecurityGroup{
			GUID:           s.GUID,
			Name:           s.Name,
			RunningDefault: s.RunningDefault,
			StagingDefault: s.StagingDefault,
		}

		var getErr error
		spaces, warnings, getErr := actor.getSecurityGroupSpacesAndAssignedLifecycles(s.GUID, includeStaging)
		allWarnings = append(allWarnings, warnings...)
		if getErr != nil {
			if _, ok := getErr.(ccerror.ResourceNotFoundError); ok {
				allWarnings = append(allWarnings, getErr.Error())
				continue
			}
			return nil, Warnings(allWarnings), getErr
		}

		if securityGroup.RunningDefault {
			secGroupOrgSpaces = append(secGroupOrgSpaces,
				SecurityGroupWithOrganizationSpaceAndLifecycle{
					SecurityGroup: &securityGroup,
					Organization:  &Organization{},
					Space:         &Space{},
					Lifecycle:     ccv2.SecurityGroupLifecycleRunning,
				})
		}

		if securityGroup.StagingDefault {
			secGroupOrgSpaces = append(secGroupOrgSpaces,
				SecurityGroupWithOrganizationSpaceAndLifecycle{
					SecurityGroup: &securityGroup,
					Organization:  &Organization{},
					Space:         &Space{},
					Lifecycle:     ccv2.SecurityGroupLifecycleStaging,
				})
		}

		if len(spaces) == 0 {
			if !securityGroup.RunningDefault && !securityGroup.StagingDefault {
				secGroupOrgSpaces = append(secGroupOrgSpaces,
					SecurityGroupWithOrganizationSpaceAndLifecycle{
						SecurityGroup: &securityGroup,
						Organization:  &Organization{},
						Space:         &Space{},
					})
			}

			continue
		}

		for _, sp := range spaces {
			space := Space{
				GUID: sp.GUID,
				Name: sp.Name,
			}

			var org Organization

			if cached, ok := cachedOrgs[sp.OrganizationGUID]; ok {
				org = cached
			} else {
				var getOrgErr error
				o, warnings, getOrgErr := actor.CloudControllerClient.GetOrganization(sp.OrganizationGUID)
				allWarnings = append(allWarnings, warnings...)
				if getOrgErr != nil {
					if _, ok := getOrgErr.(ccerror.ResourceNotFoundError); ok {
						allWarnings = append(allWarnings, getOrgErr.Error())
						continue
					}
					return nil, Warnings(allWarnings), getOrgErr
				}

				org = Organization{
					GUID: o.GUID,
					Name: o.Name,
				}
				cachedOrgs[org.GUID] = org
			}

			secGroupOrgSpaces = append(secGroupOrgSpaces,
				SecurityGroupWithOrganizationSpaceAndLifecycle{
					SecurityGroup: &securityGroup,
					Organization:  &org,
					Space:         &space,
					Lifecycle:     sp.Lifecycle,
				})
		}
	}

	// Sort the results alphabetically by security group, then org, then space
	sort.Slice(secGroupOrgSpaces,
		func(i, j int) bool {
			switch {
			case secGroupOrgSpaces[i].SecurityGroup.Name < secGroupOrgSpaces[j].SecurityGroup.Name:
				return true
			case secGroupOrgSpaces[i].SecurityGroup.Name > secGroupOrgSpaces[j].SecurityGroup.Name:
				return false
			case secGroupOrgSpaces[i].SecurityGroup.RunningDefault && !secGroupOrgSpaces[i].SecurityGroup.RunningDefault:
				return true
			case !secGroupOrgSpaces[i].SecurityGroup.RunningDefault && secGroupOrgSpaces[i].SecurityGroup.RunningDefault:
				return false
			case secGroupOrgSpaces[i].Organization.Name < secGroupOrgSpaces[j].Organization.Name:
				return true
			case secGroupOrgSpaces[i].Organization.Name > secGroupOrgSpaces[j].Organization.Name:
				return false
			case secGroupOrgSpaces[i].SecurityGroup.StagingDefault && !secGroupOrgSpaces[i].SecurityGroup.StagingDefault:
				return true
			case !secGroupOrgSpaces[i].SecurityGroup.StagingDefault && secGroupOrgSpaces[i].SecurityGroup.StagingDefault:
				return false
			case secGroupOrgSpaces[i].Space.Name < secGroupOrgSpaces[j].Space.Name:
				return true
			case secGroupOrgSpaces[i].Space.Name > secGroupOrgSpaces[j].Space.Name:
				return false
			}

			return secGroupOrgSpaces[i].Lifecycle < secGroupOrgSpaces[j].Lifecycle
		})

	return secGroupOrgSpaces, Warnings(allWarnings), nil
}

// GetSpaceRunningSecurityGroupsBySpace returns a list of all security groups
// bound to this space in the 'running' lifecycle phase.
func (actor Actor) GetSpaceRunningSecurityGroupsBySpace(spaceGUID string) ([]SecurityGroup, Warnings, error) {
	ccv2SecurityGroups, warnings, err := actor.CloudControllerClient.GetSpaceRunningSecurityGroupsBySpace(spaceGUID)
	return processSecurityGroups(spaceGUID, ccv2SecurityGroups, Warnings(warnings), err)
}

// GetSpaceStagingSecurityGroupsBySpace returns a list of all security groups
// bound to this space in the 'staging' lifecycle phase. with an optional
func (actor Actor) GetSpaceStagingSecurityGroupsBySpace(spaceGUID string) ([]SecurityGroup, Warnings, error) {
	ccv2SecurityGroups, warnings, err := actor.CloudControllerClient.GetSpaceStagingSecurityGroupsBySpace(spaceGUID)
	return processSecurityGroups(spaceGUID, ccv2SecurityGroups, Warnings(warnings), err)
}

func (actor Actor) UnbindSecurityGroupByNameAndSpace(securityGroupName string, spaceGUID string, lifecycle ccv2.SecurityGroupLifecycle) (Warnings, error) {
	if lifecycle != ccv2.SecurityGroupLifecycleRunning && lifecycle != ccv2.SecurityGroupLifecycleStaging {
		return nil, actionerror.InvalidLifecycleError{Lifecycle: lifecycle}
	}

	var allWarnings Warnings

	securityGroup, warnings, err := actor.GetSecurityGroupByName(securityGroupName)

	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.unbindSecurityGroupAndSpace(securityGroup, spaceGUID, lifecycle)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}

func (actor Actor) UnbindSecurityGroupByNameOrganizationNameAndSpaceName(securityGroupName string, orgName string, spaceName string, lifecycle ccv2.SecurityGroupLifecycle) (Warnings, error) {
	if lifecycle != ccv2.SecurityGroupLifecycleRunning && lifecycle != ccv2.SecurityGroupLifecycleStaging {
		return nil, actionerror.InvalidLifecycleError{Lifecycle: lifecycle}
	}

	var allWarnings Warnings

	securityGroup, warnings, err := actor.GetSecurityGroupByName(securityGroupName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

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

	warnings, err = actor.unbindSecurityGroupAndSpace(securityGroup, space.GUID, lifecycle)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}

func (actor Actor) unbindSecurityGroupAndSpace(securityGroup SecurityGroup, spaceGUID string, lifecycle ccv2.SecurityGroupLifecycle) (Warnings, error) {
	if lifecycle == ccv2.SecurityGroupLifecycleRunning {
		return actor.doUnbind(securityGroup, spaceGUID, lifecycle,
			actor.isRunningSecurityGroupBoundToSpace,
			actor.isStagingSecurityGroupBoundToSpace,
			actor.CloudControllerClient.RemoveSpaceFromRunningSecurityGroup)
	}

	return actor.doUnbind(securityGroup, spaceGUID, lifecycle,
		actor.isStagingSecurityGroupBoundToSpace,
		actor.isRunningSecurityGroupBoundToSpace,
		actor.CloudControllerClient.RemoveSpaceFromStagingSecurityGroup)
}

func (Actor) doUnbind(securityGroup SecurityGroup,
	spaceGUID string,
	lifecycle ccv2.SecurityGroupLifecycle,
	requestedPhaseSecurityGroupBoundToSpace func(string, string) (bool, Warnings, error),
	otherPhaseSecurityGroupBoundToSpace func(string, string) (bool, Warnings, error),
	removeSpaceFromPhaseSecurityGroup func(string, string) (ccv2.Warnings, error)) (Warnings, error) {

	requestedPhaseBound, allWarnings, err := requestedPhaseSecurityGroupBoundToSpace(securityGroup.Name, spaceGUID)
	if err != nil {
		return allWarnings, err
	}

	if !requestedPhaseBound {
		otherBound, warnings, otherr := otherPhaseSecurityGroupBoundToSpace(securityGroup.Name, spaceGUID)
		allWarnings = append(allWarnings, warnings...)

		if otherr != nil {
			return allWarnings, otherr
		} else if otherBound {
			return allWarnings, actionerror.SecurityGroupNotBoundError{Name: securityGroup.Name, Lifecycle: lifecycle}
		} else {
			return allWarnings, nil
		}
	}

	ccv2Warnings, err := removeSpaceFromPhaseSecurityGroup(securityGroup.GUID, spaceGUID)
	allWarnings = append(allWarnings, Warnings(ccv2Warnings)...)
	return allWarnings, err
}

func extractSecurityGroupRules(securityGroup SecurityGroup, lifecycle ccv2.SecurityGroupLifecycle) []SecurityGroupRule {
	securityGroupRules := make([]SecurityGroupRule, len(securityGroup.Rules))

	for i, rule := range securityGroup.Rules {
		securityGroupRules[i] = SecurityGroupRule{
			Name:        securityGroup.Name,
			Description: rule.Description,
			Destination: rule.Destination,
			Lifecycle:   lifecycle,
			Ports:       rule.Ports,
			Protocol:    rule.Protocol,
		}
	}

	return securityGroupRules
}

func processSecurityGroups(spaceGUID string, ccv2SecurityGroups []ccv2.SecurityGroup, warnings Warnings, err error) ([]SecurityGroup, Warnings, error) {
	if err != nil {
		switch err.(type) {
		case ccerror.ResourceNotFoundError:
			return []SecurityGroup{}, warnings, actionerror.SpaceNotFoundError{GUID: spaceGUID}
		default:
			return []SecurityGroup{}, warnings, err
		}
	}

	securityGroups := make([]SecurityGroup, len(ccv2SecurityGroups))
	for i, securityGroup := range ccv2SecurityGroups {
		securityGroups[i] = SecurityGroup(securityGroup)
	}

	return securityGroups, warnings, nil
}

func (actor Actor) isRunningSecurityGroupBoundToSpace(securityGroupName string, spaceGUID string) (bool, Warnings, error) {
	ccv2SecurityGroups, warnings, err := actor.CloudControllerClient.GetSpaceRunningSecurityGroupsBySpace(spaceGUID, ccv2.QQuery{
		Filter:   ccv2.NameFilter,
		Operator: ccv2.EqualOperator,
		Values:   []string{securityGroupName},
	})
	return len(ccv2SecurityGroups) > 0, Warnings(warnings), err
}

func (actor Actor) isStagingSecurityGroupBoundToSpace(securityGroupName string, spaceGUID string) (bool, Warnings, error) {
	ccv2SecurityGroups, warnings, err := actor.CloudControllerClient.GetSpaceStagingSecurityGroupsBySpace(spaceGUID, ccv2.QQuery{
		Filter:   ccv2.NameFilter,
		Operator: ccv2.EqualOperator,
		Values:   []string{securityGroupName},
	})
	return len(ccv2SecurityGroups) > 0, Warnings(warnings), err
}
