package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"strings"
)

type RouteRepository interface {
	FindAll(config *configuration.Configuration) (routes []cf.Route, err error)
	FindByHost(config *configuration.Configuration, host string) (route cf.Route, err error)
	Create(config *configuration.Configuration, newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, err error)
	Bind(config *configuration.Configuration, route cf.Route, app cf.Application) (err error)
}

type CloudControllerRouteRepository struct {
}

func (repo CloudControllerRouteRepository) FindAll(config *configuration.Configuration) (routes []cf.Route, err error) {
	path := fmt.Sprintf("%s/v2/routes?inline-relations-depth=1", config.Target)

	request, err := NewRequest("GET", path, config.AccessToken, nil)
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

func (repo CloudControllerRouteRepository) FindByHost(config *configuration.Configuration, host string) (route cf.Route, err error) {
	path := fmt.Sprintf("%s/v2/routes?q=host%s", config.Target, "%3A"+host)

	request, err := NewRequest("GET", path, config.AccessToken, nil)
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

func (repo CloudControllerRouteRepository) Create(config *configuration.Configuration, newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, err error) {
	path := fmt.Sprintf("%s/v2/routes", config.Target)
	data := fmt.Sprintf(
		`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`,
		newRoute.Host, domain.Guid, config.Space.Guid,
	)
	request, err := NewRequest("POST", path, config.AccessToken, strings.NewReader(data))
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

func (repo CloudControllerRouteRepository) Bind(config *configuration.Configuration, route cf.Route, app cf.Application) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", config.Target, app.Guid, route.Guid)
	request, err := NewRequest("PUT", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	_, err = PerformRequest(request)

	return
}
