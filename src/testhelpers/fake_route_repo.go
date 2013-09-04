package testhelpers

import (
"cf"
"cf/configuration"
	"errors"
)

type FakeRouteRepository struct {
	FindByHostHost      string
	FindByHostErr      bool
	FindByHostRoute      cf.Route

	CreatedRoute cf.Route
	CreatedRouteDomain cf.Domain

	BoundRoute cf.Route
	BoundApp cf.Application

	FindAllErr bool
	FindAllRoutes []cf.Route
}

func (repo *FakeRouteRepository) FindAll(config *configuration.Configuration) (routes []cf.Route, err error){
	if repo.FindAllErr {
		err = errors.New("Error finding all routes")
	}

	routes = repo.FindAllRoutes
	return
}

func (repo *FakeRouteRepository) FindByHost(config *configuration.Configuration, host string) (route cf.Route, err error) {
	repo.FindByHostHost = host

	if repo.FindByHostErr {
		err = errors.New("Route not found")
	}

	route = repo.FindByHostRoute
	return
}

func (repo *FakeRouteRepository) Create(config *configuration.Configuration, newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, err error){
	repo.CreatedRoute = newRoute
	repo.CreatedRouteDomain = domain

	createdRoute = cf.Route{
		Host: newRoute.Host,
		Guid: newRoute.Host + "-guid",
	}
	return
}

func (repo *FakeRouteRepository) Bind(config *configuration.Configuration, route cf.Route, app cf.Application) (err error){
	repo.BoundRoute = route
	repo.BoundApp = app
	return
}



