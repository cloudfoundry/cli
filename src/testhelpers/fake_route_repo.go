package testhelpers

import (
	"cf"
	"cf/api"
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

func (repo *FakeRouteRepository) FindAll() (routes []cf.Route, apiErr *api.ApiError) {
	if repo.FindAllErr {
		apiErr = api.NewApiErrorWithMessage("Error finding all routes")
	}

	routes = repo.FindAllRoutes
	return
}

func (repo *FakeRouteRepository) FindByHost(host string) (route cf.Route, apiErr *api.ApiError) {
	repo.FindByHostHost = host

	if repo.FindByHostErr {
		apiErr = api.NewApiErrorWithMessage("Route not found")
	}

	route = repo.FindByHostRoute
	return
}

func (repo *FakeRouteRepository) Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiErr *api.ApiError) {
	repo.CreatedRoute = newRoute
	repo.CreatedRouteDomain = domain

	createdRoute = cf.Route{
		Host: newRoute.Host,
		Guid: newRoute.Host + "-guid",
	}
	return
}

func (repo *FakeRouteRepository) Bind(route cf.Route, app cf.Application) (apiErr *api.ApiError) {
	repo.BoundRoute = route
	repo.BoundApp = app
	return
}



