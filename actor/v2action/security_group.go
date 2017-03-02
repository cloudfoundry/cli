package v2action

import "fmt"

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
	return SecurityGroup{}, Warnings{}, nil
}

func (actor Actor) BindSecurityGroupToSpace(securityGroupGUID string, spaceGUID string) (Warnings, error) {
	return Warnings{}, nil

}
