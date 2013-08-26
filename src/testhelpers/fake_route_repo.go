package testhelpers

import (
"cf"
"cf/configuration"
)

type FakeRouteRepository struct {
	CreatedRoute cf.Route
	CreatedRouteDomain cf.Domain

	BoundRoute cf.Route
	BoundApp cf.Application
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



