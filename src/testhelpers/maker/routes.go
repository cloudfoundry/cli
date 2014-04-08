package maker

import "cf/models"

var routeGuid func() string

func init() {
	routeGuid = guidGenerator("route")
}

func NewRouteFields(overrides Overrides) (route models.RouteFields) {
	route.Guid = routeGuid()
	route.Host = "route-host"

	if overrides.Has("Guid") {
		route.Guid = overrides.Get("Guid").(string)
	}

	if overrides.Has("Host") {
		route.Host = overrides.Get("Host").(string)
	}

	return
}

func NewRoute(overrides Overrides) (route models.Route) {
	route.RouteFields = NewRouteFields(overrides)
	return
}
