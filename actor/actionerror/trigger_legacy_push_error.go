package actionerror

import "fmt"

type TriggerLegacyPushError struct {
	DomainRelated      bool
	HostnameRelated    bool
	GlobalRelated      []string
	InheritanceRelated bool
	RandomRouteRelated bool
}

func (e TriggerLegacyPushError) Error() string {
	return fmt.Sprintf("Triggering legacy push due to - Domain(s): %t Hostname(s): %t Random Route: %t", e.DomainRelated, e.HostnameRelated, e.RandomRouteRelated)
}
