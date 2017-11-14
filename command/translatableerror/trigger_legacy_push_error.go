package translatableerror

type TriggerLegacyPushError struct {
	DomainRelated      bool
	HostnameRelated    bool
	RandomRouteRelated bool
}

func (TriggerLegacyPushError) LegacyMain() {}

func (e TriggerLegacyPushError) Error() string {
	if e.DomainRelated {
		return `App manifest declares routes using domain or domains attributes.
These attributes are not processed by 'v2-push' and may be deprecated in the future.
You can prevent this message by declaring routes in a "routes" section.
See http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#routes.
Continuing processing using 'push' command...
`
	}
	return ""
}

func (e TriggerLegacyPushError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
