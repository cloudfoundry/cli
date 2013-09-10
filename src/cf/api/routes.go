package api

import (
	"cf"
	"cf/configuration"
	"fmt"
	"strings"
)

type RouteRepository interface {
	FindAll() (routes []cf.Route, apiErr *ApiError)
	FindByHost(host string) (route cf.Route, apiErr *ApiError)
	Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiErr *ApiError)
	Bind(route cf.Route, app cf.Application) (apiErr *ApiError)
}

type CloudControllerRouteRepository struct {
	config    *configuration.Configuration
	apiClient ApiClient
}

func NewCloudControllerRouteRepository(config *configuration.Configuration, apiClient ApiClient) (repo CloudControllerRouteRepository) {
	repo.config = config
	repo.apiClient = apiClient
	return
}

func (repo CloudControllerRouteRepository) FindAll() (routes []cf.Route, apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/routes?inline-relations-depth=1", repo.config.Target)

	request, apiErr := NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(RoutesResponse)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, response)
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

func (repo CloudControllerRouteRepository) FindByHost(host string) (route cf.Route, apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/routes?q=host%s", repo.config.Target, "%3A"+host)

	request, apiErr := NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	response := new(ApiResponse)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, response)
	if apiErr != nil {
		return
	}

	if len(response.Resources) == 0 {
		apiErr = NewApiErrorWithMessage("Route not found")
		return
	}

	resource := response.Resources[0]
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host

	return
}

func (repo CloudControllerRouteRepository) Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/routes", repo.config.Target)
	data := fmt.Sprintf(
		`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`,
		newRoute.Host, domain.Guid, repo.config.Space.Guid,
	)
	request, apiErr := NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	resource := new(Resource)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, resource)
	if apiErr != nil {
		return
	}

	createdRoute.Guid = resource.Metadata.Guid
	createdRoute.Host = resource.Entity.Host
	return
}

func (repo CloudControllerRouteRepository) Bind(route cf.Route, app cf.Application) (apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, app.Guid, route.Guid)
	request, apiErr := NewRequest("PUT", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	apiErr = repo.apiClient.PerformRequest(request)

	return
}
