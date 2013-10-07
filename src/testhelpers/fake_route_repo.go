package testhelpers

import (
	"cf"
	"cf/net"
	"fmt"
)

type FakeRouteRepository struct {
	FindByHostHost       string
	FindByHostErr        bool
	FindByHostRoute      cf.Route

	FindByHostAndDomainHost     string
	FindByHostAndDomainDomain   string
	FindByHostAndDomainRoute    cf.Route
	FindByHostAndDomainErr      bool
	FindByHostAndDomainNotFound bool

	CreatedRoute       cf.Route
	CreatedRouteDomain cf.Domain

	CreateInSpaceRoute cf.Route
	CreateInSpaceDomain cf.Domain
	CreateInSpaceSpace cf.Space
	CreateInSpaceCreatedRoute cf.Route

	BoundRoute cf.Route
	BoundApp   cf.Application

	UnboundRoute cf.Route
	UnboundApp   cf.Application

	FindAllErr    bool
	FindAllRoutes []cf.Route
}

func (repo *FakeRouteRepository) FindAll() (routes []cf.Route, apiResponse net.ApiResponse) {
	if repo.FindAllErr {
		apiResponse = net.NewApiStatusWithMessage("Error finding all routes")
	}

	routes = repo.FindAllRoutes
	return
}

func (repo *FakeRouteRepository) FindByHost(host string) (route cf.Route, apiResponse net.ApiResponse) {
	repo.FindByHostHost = host

	if repo.FindByHostErr {
		apiResponse = net.NewApiStatusWithMessage("Route not found")
	}

	route = repo.FindByHostRoute
	return
}

func (repo *FakeRouteRepository) FindByHostAndDomain(host, domain string) (route cf.Route, apiResponse net.ApiResponse) {
	repo.FindByHostAndDomainHost = host
	repo.FindByHostAndDomainDomain = domain

	if repo.FindByHostAndDomainErr {
		apiResponse = net.NewApiStatusWithMessage("Error finding Route")
	}

	if repo.FindByHostAndDomainNotFound {
		apiResponse = net.NewNotFoundApiStatus("Org", fmt.Sprintf("%s.%s", host, domain))
	}

	route = repo.FindByHostAndDomainRoute
	return
}

func (repo *FakeRouteRepository) Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiResponse net.ApiResponse) {
	repo.CreatedRoute = newRoute
	repo.CreatedRouteDomain = domain

	createdRoute = cf.Route{
		Host: newRoute.Host,
		Guid: newRoute.Host + "-guid",
	}
	return
}

func (repo *FakeRouteRepository) CreateInSpace(newRoute cf.Route, domain cf.Domain, space cf.Space) (createdRoute cf.Route, apiResponse net.ApiResponse) {
	repo.CreateInSpaceRoute = newRoute
	repo.CreateInSpaceDomain = domain
	repo.CreateInSpaceSpace = space

	createdRoute = repo.CreateInSpaceCreatedRoute
	return
}

func (repo *FakeRouteRepository) Bind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse) {
	repo.BoundRoute = route
	repo.BoundApp = app
	return
}

func (repo *FakeRouteRepository) Unbind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse) {
	repo.UnboundRoute = route
	repo.UnboundApp = app
	return
}



