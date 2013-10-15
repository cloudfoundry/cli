package requirements

import (
	"cf"
	"cf/api"
	"cf/net"
	"cf/terminal"
)

type RouteRequirement interface {
	Requirement
	GetRoute() cf.Route
}

type routeApiRequirement struct {
	host      string
	domain    string
	ui        terminal.UI
	routeRepo api.RouteRepository
	route     cf.Route
}

func newRouteRequirement(host, domain string, ui terminal.UI, routeRepo api.RouteRepository) (req *routeApiRequirement) {
	req = new(routeApiRequirement)
	req.host = host
	req.domain = domain
	req.ui = ui
	req.routeRepo = routeRepo
	return
}

func (req *routeApiRequirement) Execute() bool {
	var apiResponse net.ApiResponse
	req.route, apiResponse = req.routeRepo.FindByHostAndDomain(req.host, req.domain)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *routeApiRequirement) GetRoute() cf.Route {
	return req.route
}
