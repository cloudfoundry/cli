package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeRouteRepository struct {
	FindByHostHost       string
	FindByHostErr        bool
	FindByHostRoute      cf.Route

	CreatedRoute       cf.Route
	CreatedRouteDomain cf.Domain

	BoundRoute cf.Route
	BoundApp   cf.Application

	FindAllErr    bool
	FindAllRoutes []cf.Route
}

func (repo *FakeRouteRepository) FindAll() (routes []cf.Route, apiErr *net.ApiError) {
	if repo.FindAllErr {
		apiErr = net.NewApiErrorWithMessage("Error finding all routes")
	}

	routes = repo.FindAllRoutes
	return
}

func (repo *FakeRouteRepository) FindByHost(host string) (route cf.Route, apiErr *net.ApiError) {
	repo.FindByHostHost = host

	if repo.FindByHostErr {
		apiErr = net.NewApiErrorWithMessage("Route not found")
	}

	route = repo.FindByHostRoute
	return
}

func (repo *FakeRouteRepository) Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiErr *net.ApiError) {
	repo.CreatedRoute = newRoute
	repo.CreatedRouteDomain = domain

	createdRoute = cf.Route{
		Host: newRoute.Host,
		Guid: newRoute.Host + "-guid",
	}
	return
}

func (repo *FakeRouteRepository) Bind(route cf.Route, app cf.Application) (apiErr *net.ApiError) {
	repo.BoundRoute = route
	repo.BoundApp = app
	return
}



