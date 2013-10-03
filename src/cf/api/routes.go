package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type RouteRepository interface {
	FindAll() (routes []cf.Route, apiStatus net.ApiStatus)
	FindByHost(host string) (route cf.Route, apiStatus net.ApiStatus)
	FindByHostAndDomain(host, domain string) (route cf.Route, apiStatus net.ApiStatus)
	Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiStatus net.ApiStatus)
	CreateInSpace(newRoute cf.Route, domain cf.Domain, space cf.Space) (createdRoute cf.Route, apiStatus net.ApiStatus)
	Bind(route cf.Route, app cf.Application) (apiStatus net.ApiStatus)
	Unbind(route cf.Route, app cf.Application) (apiStatus net.ApiStatus)
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

func (repo CloudControllerRouteRepository) FindAll() (routes []cf.Route, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/routes?inline-relations-depth=1", repo.config.Target)

	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	response := new(RoutesResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiStatus.IsError() {
		return
	}

	for _, routeResponse := range response.Routes {
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

func (repo CloudControllerRouteRepository) FindByHost(host string) (route cf.Route, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/routes?q=host%s", repo.config.Target, "%3A"+host)

	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	response := new(ApiResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiStatus.IsError() {
		return
	}

	if len(response.Resources) == 0 {
		apiStatus = net.NewApiStatusWithMessage("Route not found")
		return
	}

	resource := response.Resources[0]
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host

	return
}

func (repo CloudControllerRouteRepository) FindByHostAndDomain(host, domainName string) (route cf.Route, apiStatus net.ApiStatus) {
	domain, apiStatus := repo.domainRepo.FindByNameInCurrentSpace(domainName)
	if apiStatus.IsError() || apiStatus.IsNotFound() {
		return
	}

	path := fmt.Sprintf("%s/v2/routes?q=host%%3A%s%%3Bdomain_guid%%3A%s", repo.config.Target, host, domain.Guid)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	response := new(ApiResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiStatus.IsError() {
		return
	}

	if len(response.Resources) == 0 {
		apiStatus = net.NewNotFoundApiStatus()
		return
	}

	resource := response.Resources[0]
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host
	route.Domain = domain

	return
}

func (repo CloudControllerRouteRepository) Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiStatus net.ApiStatus) {
	return repo.CreateInSpace(newRoute, domain, repo.config.Space)
}

func (repo CloudControllerRouteRepository) CreateInSpace(newRoute cf.Route, domain cf.Domain, space cf.Space) (createdRoute cf.Route, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/routes", repo.config.Target)
	data := fmt.Sprintf(
		`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`,
		newRoute.Host, domain.Guid, space.Guid,
	)
	request, apiStatus := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiStatus.IsError() {
		return
	}

	resource := new(Resource)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiStatus.IsError() {
		return
	}

	createdRoute.Guid = resource.Metadata.Guid
	createdRoute.Host = resource.Entity.Host
	return
}

func (repo CloudControllerRouteRepository) Bind(route cf.Route, app cf.Application) (apiStatus net.ApiStatus) {
	return repo.change("PUT", route, app)
}

func (repo CloudControllerRouteRepository) Unbind(route cf.Route, app cf.Application) (apiStatus net.ApiStatus) {
	return repo.change("DELETE", route, app)
}

func (repo CloudControllerRouteRepository) change(verb string, route cf.Route, app cf.Application) (apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, app.Guid, route.Guid)
	request, apiStatus := repo.gateway.NewRequest(verb, path, repo.config.AccessToken, nil)
	if apiStatus.IsError() {
		return
	}

	apiStatus = repo.gateway.PerformRequest(request)

	return
}
