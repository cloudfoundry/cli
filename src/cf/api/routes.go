package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type RouteResource struct {
	Resource
	Entity RouteEntity
}

func (resource RouteResource) ToFields() (fields models.RouteFields) {
	fields.Guid = resource.Metadata.Guid
	fields.Host = resource.Entity.Host
	return
}
func (resource RouteResource) ToModel() (route models.Route) {
	route.RouteFields = resource.ToFields()
	route.Domain = resource.Entity.Domain.ToFields()
	route.Space = resource.Entity.Space.ToFields()
	for _, appResource := range resource.Entity.Apps {
		route.Apps = append(route.Apps, appResource.ToFields())
	}
	return
}

type RouteEntity struct {
	Host   string
	Domain DomainResource
	Space  SpaceResource
	Apps   []ApplicationResource
}

type RouteRepository interface {
	ListRoutes(cb func(models.Route) bool) (apiResponse errors.Error)
	FindByHost(host string) (route models.Route, apiResponse errors.Error)
	FindByHostAndDomain(host, domain string) (route models.Route, apiResponse errors.Error)
	Create(host, domainGuid string) (createdRoute models.Route, apiResponse errors.Error)
	CreateInSpace(host, domainGuid, spaceGuid string) (createdRoute models.Route, apiResponse errors.Error)
	Bind(routeGuid, appGuid string) (apiResponse errors.Error)
	Unbind(routeGuid, appGuid string) (apiResponse errors.Error)
	Delete(routeGuid string) (apiResponse errors.Error)
}

type CloudControllerRouteRepository struct {
	config     configuration.Reader
	gateway    net.Gateway
	domainRepo DomainRepository
}

func NewCloudControllerRouteRepository(config configuration.Reader, gateway net.Gateway, domainRepo DomainRepository) (repo CloudControllerRouteRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.domainRepo = domainRepo
	return
}

func (repo CloudControllerRouteRepository) ListRoutes(cb func(models.Route) bool) (apiResponse errors.Error) {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		fmt.Sprintf("/v2/routes?inline-relations-depth=1"),
		RouteResource{},
		func(resource interface{}) bool {
			return cb(resource.(RouteResource).ToModel())
		})
}

func (repo CloudControllerRouteRepository) FindByHost(host string) (route models.Route, apiResponse errors.Error) {
	found := false
	apiResponse = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=%s", url.QueryEscape("host:"+host)),
		RouteResource{},
		func(resource interface{}) bool {
			route = resource.(RouteResource).ToModel()
			found = true
			return false
		})

	if apiResponse == nil && !found {
		apiResponse = errors.NewNotFoundError("Route with host %s not found", host)
	}

	return
}

func (repo CloudControllerRouteRepository) FindByHostAndDomain(host, domainName string) (route models.Route, apiResponse errors.Error) {
	domain, apiResponse := repo.domainRepo.FindByName(domainName)
	if apiResponse != nil {
		return
	}

	found := false
	apiResponse = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=%s", url.QueryEscape("host:"+host+";domain_guid:"+domain.Guid)),
		RouteResource{},
		func(resource interface{}) bool {
			route = resource.(RouteResource).ToModel()
			found = true
			return false
		})

	if apiResponse == nil && !found {
		apiResponse = errors.NewNotFoundError("Route with host %s not found", host)
	}

	return
}

func (repo CloudControllerRouteRepository) Create(host, domainGuid string) (createdRoute models.Route, apiResponse errors.Error) {
	return repo.CreateInSpace(host, domainGuid, repo.config.SpaceFields().Guid)
}

func (repo CloudControllerRouteRepository) CreateInSpace(host, domainGuid, spaceGuid string) (createdRoute models.Route, apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/routes?inline-relations-depth=1", repo.config.ApiEndpoint())
	data := fmt.Sprintf(`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`, host, domainGuid, spaceGuid)

	resource := new(RouteResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)
	if apiResponse != nil {
		return
	}

	createdRoute = resource.ToModel()
	return
}

func (repo CloudControllerRouteRepository) Bind(routeGuid, appGuid string) (apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.ApiEndpoint(), appGuid, routeGuid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), nil)
}

func (repo CloudControllerRouteRepository) Unbind(routeGuid, appGuid string) (apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.ApiEndpoint(), appGuid, routeGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken())
}

func (repo CloudControllerRouteRepository) Delete(routeGuid string) (apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/routes/%s", repo.config.ApiEndpoint(), routeGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken())
}
