package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
)

// SecurityGroup represents a CF SecurityGroup.
type SecurityGroup struct {
	Name        string
	GUID        string
	Protocol    string
	Destination string
	Ports       string
	Description string
}

// SecurityGroupNotFoundError is returned when a requested security group is
// not found.
type SecurityGroupNotFoundError struct {
	Name string
}

func (e SecurityGroupNotFoundError) Error() string {
	return fmt.Sprintf("Security group '%s' not found.", e.Name)
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

func (actor Actor) BindSecurityGroupToSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.AssociateSpaceWithSecurityGroup(securityGroupGUID, spaceGUID)
	return Warnings(warnings), err
}
