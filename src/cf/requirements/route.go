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
	var apiStatus net.ApiStatus
	req.route, apiStatus = req.routeRepo.FindByHostAndDomain(req.host, req.domain)

	if apiStatus.IsError() {
		req.ui.Failed(apiStatus.Message)
		return false
	}

	if apiStatus.IsNotFound() {
		req.ui.Failed("Route not found")
		return false
	}

	return true
}

func (req *RouteApiRequirement) GetRoute() cf.Route {
	return req.route
}
