package api

import (
	"cf"
	"cf/net"
)

type FakeRouteRepository struct {
	FindByHostHost  string
	FindByHostErr   bool
	FindByHostRoute cf.Route

	FindByHostAndDomainHost     string
	FindByHostAndDomainDomain   string
	FindByHostAndDomainRoute    cf.Route
	FindByHostAndDomainErr      bool
	FindByHostAndDomainNotFound bool

	CreatedHost       string
	CreatedDomainGuid string

	CreateInSpaceHost         string
	CreateInSpaceDomainGuid   string
	CreateInSpaceSpaceGuid    string
	CreateInSpaceCreatedRoute cf.Route
	CreateInSpaceErr          bool

	BoundRouteGuid string
	BoundAppGuid   string

	UnboundRouteGuid string
	UnboundAppGuid   string

	ListErr bool
	Routes  []cf.Route

	DeleteRouteGuid string
}

func (repo *FakeRouteRepository) ListRoutes(stop chan bool) (routesChan chan []cf.Route, statusChan chan net.ApiResponse) {
	routesChan = make(chan []cf.Route, 4)
	statusChan = make(chan net.ApiResponse, 1)

	if repo.ListErr {
		statusChan <- net.NewApiResponseWithMessage("Error finding all routes")
		close(routesChan)
		close(statusChan)
		return
	}

	go func() {
		routesCount := len(repo.Routes)
		for i := 0; i < routesCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if routesCount-i > 1 {
					routesChan <- repo.Routes[i : i+2]
				} else {
					routesChan <- repo.Routes[i:]
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
		apiResponse = net.NewNotFoundApiResponse("%s %s.%s not found", "Org", host, domain)
	}

	route = repo.FindByHostAndDomainRoute
	return
}

func (repo *FakeRouteRepository) Create(host, domainGuid string) (createdRoute cf.Route, apiResponse net.ApiResponse) {
	repo.CreatedHost = host
	repo.CreatedDomainGuid = domainGuid

	createdRoute.Guid = host + "-route-guid"

	return
}

func (repo *FakeRouteRepository) CreateInSpace(host, domainGuid, spaceGuid string) (createdRoute cf.Route, apiResponse net.ApiResponse) {
	repo.CreateInSpaceHost = host
	repo.CreateInSpaceDomainGuid = domainGuid
	repo.CreateInSpaceSpaceGuid = spaceGuid

	if repo.CreateInSpaceErr {
		apiResponse = net.NewApiResponseWithMessage("Error")
	} else {
		createdRoute = repo.CreateInSpaceCreatedRoute
	}

	return
}

func (repo *FakeRouteRepository) Bind(routeGuid, appGuid string) (apiResponse net.ApiResponse) {
	repo.BoundRouteGuid = routeGuid
	repo.BoundAppGuid = appGuid
	return
}

func (repo *FakeRouteRepository) Unbind(routeGuid, appGuid string) (apiResponse net.ApiResponse) {
	repo.UnboundRouteGuid = routeGuid
	repo.UnboundAppGuid = appGuid
	return
}

func (repo *FakeRouteRepository) Delete(routeGuid string) (apiResponse net.ApiResponse) {
	repo.DeleteRouteGuid = routeGuid
	return
}
