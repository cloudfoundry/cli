package actionerror

import "fmt"

type OrganizationQuotaNotFoundError struct {
	GUID string
}

func (e OrganizationQuotaNotFoundError) Error() string {
	return fmt.Sprintf("Organization quota with GUID '%s' not found.", e.GUID)
}
