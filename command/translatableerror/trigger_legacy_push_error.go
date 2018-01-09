package translatableerror

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

func (TriggerLegacyPushError) LegacyMain() {}

func (e TriggerLegacyPushError) Error() string {
	switch {
	case len(e.DomainHostRelated) > 0:
		return fmt.Sprintf(`Deprecation warning: Route component attributes 'domain', 'domains', 'host', 'hosts' and 'no-hostname' are deprecated. Found: %s.
Please see http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated for the currently supported syntax and other app manifest deprecations. This feature will be removed in the future.
`, strings.Join(e.DomainHostRelated, ", "))
	case len(e.GlobalRelated) > 0:
		return fmt.Sprintf(`Deprecation warning: Specifying app manifest attributes at the top level is deprecated. Found: %s.
Please see http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated for alternatives and other app manifest deprecations. This feature will be removed in the future.
`, strings.Join(e.GlobalRelated, ", "))

	case e.InheritanceRelated:
		return `Deprecation warning: App manifest inheritance is deprecated.
Please see http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated and other app manifest deprecations. This feature will be removed in the future.
`

	default:
		return ""
	}
}

func (e TriggerLegacyPushError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
