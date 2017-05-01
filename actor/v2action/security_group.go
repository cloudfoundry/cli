package v2action

import (
	"fmt"
	"sort"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// SecurityGroup represents a CF SecurityGroup.
type SecurityGroup ccv2.SecurityGroup

// SecurityGroupWithOrganizationAndSpace represents a security group with
// organization and space information.
type SecurityGroupWithOrganizationAndSpace struct {
	SecurityGroup *SecurityGroup
	Organization  *Organization
	Space         *Space
}

// SecurityGroupNotFoundError is returned when a requested security group is
// not found.
type SecurityGroupNotFoundError struct {
	Name string
}

func (e SecurityGroupNotFoundError) Error() string {
	return fmt.Sprintf("Security group '%s' not found.", e.Name)
}

func (actor Actor) BindSecurityGroupToSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.AssociateSpaceWithSecurityGroup(securityGroupGUID, spaceGUID)
	return Warnings(warnings), err
}

func (actor Actor) GetSecurityGroupByName(securityGroupName string) (SecurityGroup, Warnings, error) {
	securityGroups, warnings, err := actor.CloudControllerClient.GetSecurityGroups([]ccv2.Query{
		{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.EqualOperator,
			Value:    securityGroupName,
		},
	})

	if err != nil {
		return SecurityGroup{}, Warnings(warnings), err
	}

	if len(securityGroups) == 0 {
		return SecurityGroup{}, Warnings(warnings), SecurityGroupNotFoundError{securityGroupName}
	}

	securityGroup := SecurityGroup{
		Name: securityGroups[0].Name,
		GUID: securityGroups[0].GUID,
	}
	return securityGroup, Warnings(warnings), nil
}

// GetSecurityGroupsWithOrganizationAndSpace returns a list of security groups
// with org and space information.
func (actor Actor) GetSecurityGroupsWithOrganizationAndSpace() ([]SecurityGroupWithOrganizationAndSpace, Warnings, error) {
	var err error

	securityGroups, allWarnings, err := actor.CloudControllerClient.GetSecurityGroups(nil)
	if err != nil {
		return nil, Warnings(allWarnings), err
	}

	cachedOrgs := make(map[string]Organization)

	var secGroupOrgSpaces []SecurityGroupWithOrganizationAndSpace

	for _, s := range securityGroups {
		securityGroup := SecurityGroup{
			GUID: s.GUID,
			Name: s.Name,
		}

		spaces, warnings, err := actor.CloudControllerClient.GetSpacesBySecurityGroup(s.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, Warnings(allWarnings), err
		}

		if len(spaces) == 0 {
			secGroupOrgSpaces = append(secGroupOrgSpaces,
				SecurityGroupWithOrganizationAndSpace{
					SecurityGroup: &securityGroup,
					Organization:  &Organization{},
					Space:         &Space{},
				})
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
				o, warnings, err := actor.CloudControllerClient.GetOrganization(sp.OrganizationGUID)
				allWarnings = append(allWarnings, warnings...)
				if err != nil {
					return nil, Warnings(allWarnings), err
				}

				org = Organization{
					GUID: o.GUID,
					Name: o.Name,
				}
				cachedOrgs[org.GUID] = org
			}

			secGroupOrgSpaces = append(secGroupOrgSpaces,
				SecurityGroupWithOrganizationAndSpace{
					SecurityGroup: &securityGroup,
					Organization:  &org,
					Space:         &space,
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
			case secGroupOrgSpaces[i].Organization.Name < secGroupOrgSpaces[j].Organization.Name:
				return true
			case secGroupOrgSpaces[i].Organization.Name > secGroupOrgSpaces[j].Organization.Name:
				return false
			}

			return secGroupOrgSpaces[i].Space.Name < secGroupOrgSpaces[j].Space.Name
		})
	return secGroupOrgSpaces, Warnings(allWarnings), err
}

// GetDomain returns the shared or private domain associated with the provided
// Domain GUID.
func (actor Actor) GetSpaceRunningSecurityGroupsBySpace(spaceGUID string) ([]SecurityGroup, Warnings, error) {
	ccv2SecurityGroups, warnings, err := actor.CloudControllerClient.GetSpaceRunningSecurityGroupsBySpace(spaceGUID)
	return processSecurityGroups(spaceGUID, ccv2SecurityGroups, Warnings(warnings), err)
}

func (actor Actor) GetSpaceStagingSecurityGroupsBySpace(spaceGUID string) ([]SecurityGroup, Warnings, error) {
	ccv2SecurityGroups, warnings, err := actor.CloudControllerClient.GetSpaceStagingSecurityGroupsBySpace(spaceGUID)
	return processSecurityGroups(spaceGUID, ccv2SecurityGroups, Warnings(warnings), err)
}

func (actor Actor) UnbindSecurityGroupByNameAndSpace(securityGroupName string, spaceGUID string) (Warnings, error) {
	var allWarnings Warnings

	securityGroup, warnings, err := actor.GetSecurityGroupByName(securityGroupName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.unbindSecurityGroupAndSpace(securityGroup.GUID, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}

func (actor Actor) UnbindSecurityGroupByNameOrganizationNameAndSpaceName(securityGroupName string, orgName string, spaceName string) (Warnings, error) {
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

	warnings, err = actor.unbindSecurityGroupAndSpace(securityGroup.GUID, space.GUID)
	allWarnings = append(allWarnings, warnings...)
	return allWarnings, err
}

func (actor Actor) unbindSecurityGroupAndSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.RemoveSpaceFromSecurityGroup(securityGroupGUID, spaceGUID)
	return Warnings(warnings), err
}

func extractSecurityGroupRules(securityGroup SecurityGroup, lifecycle string) []SecurityGroupRule {
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
			return []SecurityGroup{}, warnings, SpaceNotFoundError{GUID: spaceGUID}
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
