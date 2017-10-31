package actionerror

import "fmt"

// MultipleSpacesFoundError represents the scenario when the cloud
// controller returns multiple spaces when filtering by name. This is a
// far out edge case and should not happen.
type MultipleSpacesFoundError struct {
	Name    string
	OrgGUID string
}

func (e MultipleSpacesFoundError) Error() string {
	return fmt.Sprintf("Multiple spaces found matching organization GUID '%s' and name '%s'", e.OrgGUID, e.Name)
}
