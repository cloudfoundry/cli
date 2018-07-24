package actionerror

import (
	"fmt"
	"strings"
)

type MultipleOrganizationQuotasFoundForNameError struct {
	Name  string
	GUIDs []string
}

func (e MultipleOrganizationQuotasFoundForNameError) Error() string {
	return fmt.Sprintf("Organization quota name '%s' references multiple quotas. GUIDs: '%s'", e.Name, strings.Join(e.GUIDs, ", "))
}
