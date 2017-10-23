package actionerror

import "fmt"

// NoDomainsFoundError is returned when there are no private or shared domains
// accessible to an organization.
type NoDomainsFoundError struct {
	OrganizationGUID string
}

func (e NoDomainsFoundError) Error() string {
	return fmt.Sprintf("No private or shared domains found for organization (GUID: %s)", e.OrganizationGUID)
}
