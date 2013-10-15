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
}

type RouteResource struct {
	Resource
	Entity RouteEntity
}

type RouteEntity struct {
	Host   string
	Domain Resource
	Apps   []Resource
}

type RouteRepository interface {
	FindAll() (routes []cf.Route, apiResponse net.ApiResponse)
	FindByHost(host string) (route cf.Route, apiResponse net.ApiResponse)
	FindByHostAndDomain(host, domain string) (route cf.Route, apiResponse net.ApiResponse)
	Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiResponse net.ApiResponse)
	CreateInSpace(newRoute cf.Route, domain cf.Domain, space cf.Space) (createdRoute cf.Route, apiResponse net.ApiResponse)
	Bind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse)
	Unbind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse)
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

func (repo CloudControllerRouteRepository) FindAll() (routes []cf.Route, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/routes?inline-relations-depth=1", repo.config.Target)

	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	routesResources := new(PaginatedRouteResources)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, routesResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, routeResponse := range routesResources.Resources {
		domainResource := routeResponse.Entity.Domain
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
				AppNames: appNames,
			},
		)
	}
	return
}

func (repo CloudControllerRouteRepository) FindByHost(host string) (route cf.Route, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/routes?q=host%s", repo.config.Target, "%3A"+host)

	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedRouteResources)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(resources.Resources) == 0 {
		apiResponse = net.NewApiResponseWithMessage("Route not found")
		return
	}

	resource := resources.Resources[0]
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host

	return
}

func (repo CloudControllerRouteRepository) FindByHostAndDomain(host, domainName string) (route cf.Route, apiResponse net.ApiResponse) {
	domain, apiResponse := repo.domainRepo.FindByNameInCurrentSpace(domainName)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/routes?q=host%%3A%s%%3Bdomain_guid%%3A%s", repo.config.Target, host, domain.Guid)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedRouteResources)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(resources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Route", fmt.Sprintf("%s.%s", host, domainName))
		return
	}

	resource := resources.Resources[0]
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host
	route.Domain = domain

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
	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiResponse.IsNotSuccessful() {
		return
	}

	resource := new(RouteResource)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdRoute.Guid = resource.Metadata.Guid
	createdRoute.Host = resource.Entity.Host
	return
}

func (repo CloudControllerRouteRepository) Bind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse) {
	return repo.change("PUT", route, app)
}

func (repo CloudControllerRouteRepository) Unbind(route cf.Route, app cf.Application) (apiResponse net.ApiResponse) {
	return repo.change("DELETE", route, app)
}

func (repo CloudControllerRouteRepository) change(verb string, route cf.Route, app cf.Application) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, app.Guid, route.Guid)
	request, apiResponse := repo.gateway.NewRequest(verb, path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)

	return
}
