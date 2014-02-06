package maker

import "cf/models"

var routeGuid func() string

func init() {
	routeGuid = guidGenerator("route")
}

func NewRouteFields(overrides Overrides) (route models.RouteFields) {
	route.Guid = routeGuid()
	route.Host = "route-host"

	if overrides.Has("guid") {
		route.Guid = overrides.Get("guid").(string)
	}

	if overrides.Has("host") {
		route.Host = overrides.Get("host").(string)
	}

	return
}

func NewRoute(overrides Overrides) (route models.Route) {
	route.RouteFields = NewRouteFields(overrides)
	return
}
