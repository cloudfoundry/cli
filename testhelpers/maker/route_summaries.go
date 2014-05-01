package maker

import "github.com/cloudfoundry/cli/cf/models"

var routeSummaryGuid func() string

func init() {
	routeSummaryGuid = guidGenerator("route-summary")
}

func NewRouteSummary(overrides Overrides) (routeSummary models.RouteSummary) {
	routeSummary.Guid = routeSummaryGuid()
	routeSummary.Host = "route-host"

	if overrides.Has("Guid") {
		routeSummary.Guid = overrides.Get("Guid").(string)
	}

	if overrides.Has("Host") {
		routeSummary.Host = overrides.Get("Host").(string)
	}

	return
}
