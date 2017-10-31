package actionerror

import (
	"fmt"
	"strings"
)

// MultipleOrganizationsFoundError represents the scenario when the cloud
// controller returns multiple organizations when filtering by name. This is a
// far out edge case and should not happen.
type MultipleOrganizationsFoundError struct {
	Name  string
	GUIDs []string
}

func (e MultipleOrganizationsFoundError) Error() string {
	guids := strings.Join(e.GUIDs, ", ")
	return fmt.Sprintf("Organization name '%s' matches multiple GUIDs: %s", e.Name, guids)
}
