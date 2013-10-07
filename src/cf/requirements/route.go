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

type RouteApiRequirement struct {
	host      string
	domain    string
	ui        terminal.UI
	routeRepo api.RouteRepository
	route     cf.Route
}

func NewRouteRequirement(host, domain string, ui terminal.UI, routeRepo api.RouteRepository) (req *RouteApiRequirement) {
	req = new(RouteApiRequirement)
	req.host = host
	req.domain = domain
	req.ui = ui
	req.routeRepo = routeRepo
	return
}

func (req *RouteApiRequirement) Execute() bool {
	var apiResponse net.ApiResponse
	req.route, apiResponse = req.routeRepo.FindByHostAndDomain(req.host, req.domain)

	if apiResponse.IsNotSuccessful() {
		req.ui.Failed(apiResponse.Message)
		return false
	}

	return true
}

func (req *RouteApiRequirement) GetRoute() cf.Route {
	return req.route
}
