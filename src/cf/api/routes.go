package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type RouteRepository interface {
	FindAll() (routes []cf.Route, apiErr *net.ApiError)
	FindByHost(host string) (route cf.Route, apiErr *net.ApiError)
	Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiErr *net.ApiError)
	Bind(route cf.Route, app cf.Application) (apiErr *net.ApiError)
}

type CloudControllerRouteRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerRouteRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerRouteRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerRouteRepository) FindAll() (routes []cf.Route, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/routes?inline-relations-depth=1", repo.config.Target)

	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(RoutesResponse)
	apiErr = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiErr != nil {
		return
	}

	for _, routeResponse := range response.Routes {
		routes = append(routes,
			cf.Route{
				Host: routeResponse.Entity.Host,
				Guid: routeResponse.Metadata.Guid,
				Domain: cf.Domain{
					Name: routeResponse.Entity.Domain.Entity.Name,
					Guid: routeResponse.Entity.Domain.Metadata.Guid,
				},
			},
		)
	}
	return
}

func (repo CloudControllerRouteRepository) FindByHost(host string) (route cf.Route, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/routes?q=host%s", repo.config.Target, "%3A"+host)

	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(ApiResponse)
	apiErr = repo.gateway.PerformRequestForJSONResponse(request, response)
	if apiErr != nil {
		return
	}

	if len(response.Resources) == 0 {
		apiErr = net.NewApiErrorWithMessage("Route not found")
		return
	}

	resource := response.Resources[0]
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host

	return
}

func (repo CloudControllerRouteRepository) Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/routes", repo.config.Target)
	data := fmt.Sprintf(
		`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`,
		newRoute.Host, domain.Guid, repo.config.Space.Guid,
	)
	request, apiErr := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	resource := new(Resource)
	apiErr = repo.gateway.PerformRequestForJSONResponse(request, resource)
	if apiErr != nil {
		return
	}

	createdRoute.Guid = resource.Metadata.Guid
	createdRoute.Host = resource.Entity.Host
	return
}

func (repo CloudControllerRouteRepository) Bind(route cf.Route, app cf.Application) (apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, app.Guid, route.Guid)
	request, apiErr := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	apiErr = repo.gateway.PerformRequest(request)

	return
}
