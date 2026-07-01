package actionerror

import "fmt"

// DomainNotEnforcingRoutePoliciesError is returned when a user attempts to
// add a route policy to a domain that does not have enforce_route_policies enabled.
type DomainNotEnforcingRoutePoliciesError struct {
	Name string
}

func (e DomainNotEnforcingRoutePoliciesError) Error() string {
	return fmt.Sprintf("Domain '%s' does not have route policy enforcement enabled.", e.Name)
}
