package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type PaginatedRouteResources struct {
	Resources []RouteResource `json:"resources"`
	NextUrl   string          `json:"next_url"`
}

type RouteResource struct {
	Resource
	Entity RouteEntity
}

type RouteEntity struct {
	Host   string
	Domain DomainResource
	Space  SpaceResource
	Apps   []Resource
}

type RouteRepository interface {
	ListRoutes(stop chan bool) (routesChan chan []cf.Route, statusChan chan net.ApiResponse)
	FindByHost(host string) (route cf.Route, apiResponse net.ApiResponse)
	FindByHostAndDomain(host, domain string) (route cf.Route, apiResponse net.ApiResponse)
	Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiResponse net.ApiResponse)
	CreateInSpace(newRoute cf.Route, domain cf.Domain, space cf.Space) (createdRoute cf.Route, apiResponse net.ApiResponse)
	Bind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse)
	Unbind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse)
	Delete(route cf.Route) (apiResponse net.ApiResponse)
}

type CloudControllerRouteRepository struct {
	config     *configuration.Configuration
	gateway    net.Gateway
	domainRepo DomainRepository
}

func NewCloudControllerRouteRepository(config *configuration.Configuration, gateway net.Gateway, domainRepo DomainRepository) (repo CloudControllerRouteRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.domainRepo = domainRepo
	return
}

func (repo CloudControllerRouteRepository) ListRoutes(stop chan bool) (routesChan chan []cf.Route, statusChan chan net.ApiResponse) {
	routesChan = make(chan []cf.Route, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		path := fmt.Sprintf("/v2/routes?inline-relations-depth=1")

	loop:
		for path != "" {
			select {
			case <-stop:
				break loop
			default:
				var (
					routes      []cf.Route
					apiResponse net.ApiResponse
				)
				routes, path, apiResponse = repo.findNextWithPath(path)
				if apiResponse.IsNotSuccessful() {
					statusChan <- apiResponse
					close(routesChan)
					close(statusChan)
					return
				}

				if len(routes) > 0 {
					routesChan <- routes
				}
			}
		}
		close(routesChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerRouteRepository) FindByHost(host string) (route cf.Route, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=host%s", "%3A"+host)
	return repo.findOneWithPath(path)
}

func (repo CloudControllerRouteRepository) FindByHostAndDomain(host, domainName string) (route cf.Route, apiResponse net.ApiResponse) {
	domain, apiResponse := repo.domainRepo.FindByName(domainName)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=host%%3A%s%%3Bdomain_guid%%3A%s", host, domain.Guid)
	route, apiResponse = repo.findOneWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	route.Domain = domain
	return
}

func (repo CloudControllerRouteRepository) findOneWithPath(path string) (route cf.Route, apiResponse net.ApiResponse) {
	routes, _, apiResponse := repo.findNextWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(routes) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Route not found")
		return
	}

	route = routes[0]
	return
}

func (repo CloudControllerRouteRepository) findNextWithPath(path string) (routes []cf.Route, nextUrl string, apiResponse net.ApiResponse) {
	routesResources := new(PaginatedRouteResources)
	apiResponse = repo.gateway.GetResource(repo.config.Target+path, repo.config.AccessToken, routesResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = routesResources.NextUrl

	for _, routeResponse := range routesResources.Resources {
		domainResource := routeResponse.Entity.Domain
		spaceResource := routeResponse.Entity.Space
		appNames := []string{}

		for _, appResource := range routeResponse.Entity.Apps {
			appNames = append(appNames, appResource.Entity.Name)
		}

		routes = append(routes,
			cf.Route{
				Host: routeResponse.Entity.Host,
				Guid: routeResponse.Metadata.Guid,
				Domain: cf.Domain{
					Name: domainResource.Entity.Name,
					Guid: domainResource.Metadata.Guid,
				},
				Space: cf.Space{
					Name: spaceResource.Entity.Name,
					Guid: spaceResource.Metadata.Guid,
				},
				AppNames: appNames,
			},
		)
	}
	return
}

func (repo CloudControllerRouteRepository) Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiResponse net.ApiResponse) {
	return repo.CreateInSpace(newRoute, domain, repo.config.Space)
}

func (repo CloudControllerRouteRepository) CreateInSpace(newRoute cf.Route, domain cf.Domain, space cf.Space) (createdRoute cf.Route, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/routes", repo.config.Target)
	data := fmt.Sprintf(
		`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`,
		newRoute.Host, domain.Guid, space.Guid,
	)

	resource := new(RouteResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdRoute.Guid = resource.Metadata.Guid
	createdRoute.Host = resource.Entity.Host
	createdRoute.Domain = domain

	return
}

func (repo CloudControllerRouteRepository) Bind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, app.Guid, route.Guid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, nil)
}

func (repo CloudControllerRouteRepository) Unbind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, app.Guid, route.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerRouteRepository) Delete(route cf.Route) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/routes/%s", repo.config.Target, route.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
