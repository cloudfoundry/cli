package actionerror

import "fmt"

type OrganizationQuotaNotFoundForNameError struct {
	Name string
}

func (e OrganizationQuotaNotFoundForNameError) Error() string {
	return fmt.Sprintf("Organization quota with name '%s' not found.", e.Name)
}
