package maker

import "cf"

var routeSummaryGuid func () string

func init() {
	routeSummaryGuid = guidGenerator("route-summary")
}

func NewRouteSummary(overrides Overrides) (routeSummary cf.RouteSummary) {
	routeSummary.Guid = routeSummaryGuid()
	routeSummary.Host = "route-host"

	if overrides.Has("guid") {
		routeSummary.Guid = overrides.Get("guid").(string)
	}

	if overrides.Has("host") {
		routeSummary.Host = overrides.Get("host").(string)
	}

	return
}
