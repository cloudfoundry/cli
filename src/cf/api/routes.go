package api

import (
	"cf"
	"cf/configuration"
	"fmt"
	"strings"
)

type RouteRepository interface {
	Create(config *configuration.Configuration, newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, err error)
	Bind(config *configuration.Configuration, route cf.Route, app cf.Application) (err error)
}

type CloudControllerRouteRepository struct {
}

func (repo CloudControllerRouteRepository) Create(config *configuration.Configuration, newRoute cf.Route, domain cf.Domain) (createdRoute cf.Route, err error) {
	path := fmt.Sprintf("%s/v2/routes", config.Target)
	data := fmt.Sprintf(
		`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`,
		newRoute.Host, domain.Guid, config.Space.Guid,
	)
	request, err := NewAuthorizedRequest("POST", path, config.AccessToken, strings.NewReader(data))
	if err != nil {
		return
	}

	resource := new(Resource)
	err = PerformRequestForBody(request, resource)
	if err != nil {
		return
	}

	createdRoute.Guid = resource.Metadata.Guid
	createdRoute.Host = resource.Entity.Host
	return
}

func (repo CloudControllerRouteRepository) Bind(config *configuration.Configuration, route cf.Route, app cf.Application) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", config.Target, app.Guid, route.Guid)
	request, err := NewAuthorizedRequest("PUT", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	err = PerformRequest(request)

	return
}
