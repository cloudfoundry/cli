package actionerror

import (
	"fmt"
	"strings"
)

type TriggerLegacyPushError struct {
	DomainHostRelated  []string
	GlobalRelated      []string
	InheritanceRelated bool
	RandomRouteRelated bool
}

func (e TriggerLegacyPushError) Error() string {
	if len(e.DomainHostRelated) > 0 || len(e.GlobalRelated) > 0 {
		return fmt.Sprintf("Triggering legacy push due to - Inheritance: %t, Random Route: %t, and Found: %s", e.InheritanceRelated, e.RandomRouteRelated, strings.Join(append(e.DomainHostRelated, e.GlobalRelated...), ", "))
	}
	return fmt.Sprintf("Triggering legacy push due to - Inheritance: %t and Random Route: %t", e.InheritanceRelated, e.RandomRouteRelated)
}
