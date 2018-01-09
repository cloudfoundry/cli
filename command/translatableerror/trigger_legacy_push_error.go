package translatableerror

import (
	"fmt"
	"strings"
)

type TriggerLegacyPushError struct {
	DomainRelated      bool
	HostnameRelated    bool
	GlobalRelated      []string
	InheritanceRelated bool
	RandomRouteRelated bool
}

func (TriggerLegacyPushError) LegacyMain() {}

func (e TriggerLegacyPushError) Error() string {
	switch {
	case e.DomainRelated:
		return `App manifest declares routes using 'domain' or 'domains' attributes.
These attributes are not processed by 'v2-push' and may be deprecated in the future.
You can prevent this message by declaring routes in a 'routes' section.
See http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#routes.
Continuing processing using 'push' command...
`
	case e.HostnameRelated:
		return `App manifest declares routes using 'host', 'hosts', or 'no-hostname' attributes.
These attributes are not processed by 'v2-push'.
Continuing processing using 'push' command...
`
	case len(e.GlobalRelated) > 0:
		return fmt.Sprintf(`App manifest has attributes promoted to the top level. Found: %s.
Promoted attributes are not processed by 'v2-push' and may be deprecated in the future.
Continuing processing using 'push' command...`, strings.Join(e.GlobalRelated, ", "))

	case e.InheritanceRelated:
		return `Deprecation warning: App manifest inheritance is deprecated.
Please see http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated-app-manifest-features for details and other app manifest deprecations. This feature will be removed in the future.
`

	default:
		return ""
	}
}

func (e TriggerLegacyPushError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
