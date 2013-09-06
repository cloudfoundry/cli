package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"strings"
)

type RouteRepository interface {
	FindAll() (routes []cf.Route, err error)
	FindByHost(host string) (route cf.Route, err error)
	Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, err error)
	Bind(route cf.Route, app cf.Application) (err error)
}

type CloudControllerRouteRepository struct {
	config *configuration.Configuration
}

func NewCloudControllerRouteRepository(config *configuration.Configuration) (repo CloudControllerRouteRepository) {
	repo.config = config
	return
}

func (repo CloudControllerRouteRepository) FindAll() (routes []cf.Route, err error) {
	path := fmt.Sprintf("%s/v2/routes?inline-relations-depth=1", repo.config.Target)

	request, err := NewRequest("GET", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	response := new(RoutesResponse)
	_, err = PerformRequestAndParseResponse(request, response)
	if err != nil {
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

func (repo CloudControllerRouteRepository) FindByHost(host string) (route cf.Route, err error) {
	path := fmt.Sprintf("%s/v2/routes?q=host%s", repo.config.Target, "%3A"+host)

	request, err := NewRequest("GET", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	response := new(ApiResponse)
	_, err = PerformRequestAndParseResponse(request, response)
	if err != nil {
		return
	}

	if len(response.Resources) == 0 {
		err = errors.New("Route not found")
		return
	}

	resource := response.Resources[0]
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host

	return
}

func (repo CloudControllerRouteRepository) Create(newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, err error) {
	path := fmt.Sprintf("%s/v2/routes", repo.config.Target)
	data := fmt.Sprintf(
		`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`,
		newRoute.Host, domain.Guid, repo.config.Space.Guid,
	)
	request, err := NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if err != nil {
		return
	}

	resource := new(Resource)
	_, err = PerformRequestAndParseResponse(request, resource)
	if err != nil {
		return
	}

	createdRoute.Guid = resource.Metadata.Guid
	createdRoute.Host = resource.Entity.Host
	return
}

func (repo CloudControllerRouteRepository) Bind(route cf.Route, app cf.Application) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, app.Guid, route.Guid)
	request, err := NewRequest("PUT", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	_, err = PerformRequest(request)

	return
}
