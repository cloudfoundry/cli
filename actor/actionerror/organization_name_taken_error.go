package actionerror

import "fmt"

// OrganizationNameTakenError represents the error that occurs when creating
// an organization fails because an organization with that name already exists
type OrganizationNameTakenError struct {
	Name string
}

func (e OrganizationNameTakenError) Error() string {
	return fmt.Sprintf("An organization with name '%s' already exists", e.Name)
}
