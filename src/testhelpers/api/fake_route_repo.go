package api

import (
	"cf"
	"cf/net"
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
	CreateInSpaceErr bool

	BoundRoute cf.Route
	BoundApp   cf.Application

	UnboundRoute cf.Route
	UnboundApp   cf.Application

	FindAllErr    bool
	AllRoutes []cf.Route

	DeleteRoute cf.Route
}

func (repo *FakeRouteRepository) ListRoutes(stop chan bool) (routesChan chan []cf.Route, statusChan chan net.ApiResponse) {
	routesChan = make(chan []cf.Route, 4)
	statusChan = make(chan net.ApiResponse, 1)

	if repo.FindAllErr {
		statusChan <- net.NewApiResponseWithMessage("Error finding all routes")
		close(routesChan)
		close(statusChan)
		return
	}


	go func() {
		routesCount := len(repo.AllRoutes)
		for i:= 0; i < routesCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if routesCount - i > 1 {
					routesChan <- repo.AllRoutes[i:i+2]
				} else {
					routesChan <- repo.AllRoutes[i:]
				}
			}
		}

		close(routesChan)
		close(statusChan)

		cf.WaitForClose(stop)
	}()

	return
}

func (repo *FakeRouteRepository) FindByHost(host string) (route cf.Route, apiResponse net.ApiResponse) {
	repo.FindByHostHost = host

	if repo.FindByHostErr {
		apiResponse = net.NewApiResponseWithMessage("Route not found")
	}

	route = repo.FindByHostRoute
	return
}

func (repo *FakeRouteRepository) FindByHostAndDomain(host, domain string) (route cf.Route, apiResponse net.ApiResponse) {
	repo.FindByHostAndDomainHost = host
	repo.FindByHostAndDomainDomain = domain

	if repo.FindByHostAndDomainErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding Route")
	}

	if repo.FindByHostAndDomainNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s.%s not found","Org",host, domain)
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
		Domain: domain,
	}
	return
}

func (repo *FakeRouteRepository) CreateInSpace(newRoute cf.Route, domain cf.Domain, space cf.Space) (createdRoute cf.Route, apiResponse net.ApiResponse) {
	repo.CreateInSpaceRoute = newRoute
	repo.CreateInSpaceDomain = domain
	repo.CreateInSpaceSpace = space

	if repo.CreateInSpaceErr {
		apiResponse = net.NewApiResponseWithMessage("Error")
	} else {
		createdRoute = repo.CreateInSpaceCreatedRoute
	}

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

func (repo *FakeRouteRepository) Delete(route cf.Route) (apiResponse net.ApiResponse) {
	repo.DeleteRoute = route
	return
}
