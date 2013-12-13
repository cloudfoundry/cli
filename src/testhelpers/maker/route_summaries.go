package maker

import "cf"

var routeGuid func () string

func init() {
	routeGuid = guidGenerator("route-summary")
}

func NewRouteSummary(overrides Overrides) (routeSummary cf.RouteSummary) {
	routeSummary.Guid = routeGuid()
	routeSummary.Host = "route-host"

	guid, ok := overrides["guid"]
	if ok {
		routeSummary.Guid = guid.(string)
	}

	host, ok := overrides["host"]
	if ok {
		routeSummary.Host = host.(string)
	}

	return
}
