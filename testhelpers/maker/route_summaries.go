package maker

import "github.com/cloudfoundry/cli/cf/models"

var routeSummaryGUID func() string

func init() {
	routeSummaryGUID = guidGenerator("route-summary")
}

func NewRouteSummary(overrides Overrides) (routeSummary models.RouteSummary) {
	routeSummary.GUID = routeSummaryGUID()
	routeSummary.Host = "route-host"

	if overrides.Has("GUID") {
		routeSummary.GUID = overrides.Get("GUID").(string)
	}

	if overrides.Has("Host") {
		routeSummary.Host = overrides.Get("Host").(string)
	}

	return
}
