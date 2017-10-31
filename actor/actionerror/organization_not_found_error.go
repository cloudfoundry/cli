package actionerror

import "fmt"

// OrganizationNotFoundError represents the error that occurs when the
// organization is not found.
type OrganizationNotFoundError struct {
	GUID string
	Name string
}

func (e OrganizationNotFoundError) Error() string {
	return fmt.Sprintf("Organization '%s' not found.", e.Name)
}
