package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type RouteRepository interface {
	ListRoutes(cb func(models.Route) bool) (apiErr error)
	ListAllRoutes(cb func(models.Route) bool) (apiErr error)
	FindByHostAndDomain(host string, domain models.DomainFields) (route models.Route, apiErr error)
	Create(host string, domain models.DomainFields) (createdRoute models.Route, apiErr error)
	CheckIfExists(host string, domain models.DomainFields) (found bool, apiErr error)
	CreateInSpace(host, path, domainGuid, spaceGuid string) (createdRoute models.Route, apiErr error)
	Bind(routeGuid, appGuid string) (apiErr error)
	Unbind(routeGuid, appGuid string) (apiErr error)
	Delete(routeGuid string) (apiErr error)
}

type CloudControllerRouteRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerRouteRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerRouteRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerRouteRepository) ListRoutes(cb func(models.Route) bool) (apiErr error) {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/spaces/%s/routes?inline-relations-depth=1", repo.config.SpaceFields().Guid),
		resources.RouteResource{},
		func(resource interface{}) bool {
			return cb(resource.(resources.RouteResource).ToModel())
		})
}

func (repo CloudControllerRouteRepository) ListAllRoutes(cb func(models.Route) bool) (apiErr error) {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/routes?q=organization_guid:%s&inline-relations-depth=1", repo.config.OrganizationFields().Guid),
		resources.RouteResource{},
		func(resource interface{}) bool {
			return cb(resource.(resources.RouteResource).ToModel())
		})
}
func (repo CloudControllerRouteRepository) FindByHostAndDomain(host string, domain models.DomainFields) (route models.Route, apiErr error) {
	found := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=%s", url.QueryEscape("host:"+host+";domain_guid:"+domain.Guid)),
		resources.RouteResource{},
		func(resource interface{}) bool {
			route = resource.(resources.RouteResource).ToModel()
			found = true
			return false
		})

	if apiErr == nil && !found {
		apiErr = errors.NewModelNotFoundError("Route", host)
	}

	return
}

func (repo CloudControllerRouteRepository) Create(host string, domain models.DomainFields) (createdRoute models.Route, apiErr error) {
	return repo.CreateInSpace(host, "", domain.Guid, repo.config.SpaceFields().Guid)
}

func (repo CloudControllerRouteRepository) CheckIfExists(host string, domain models.DomainFields) (found bool, apiErr error) {
	var raw_response interface{}
	apiErr = repo.gateway.GetResource(fmt.Sprintf("%s/v2/routes/reserved/domain/%s/host/%s", repo.config.ApiEndpoint(), domain.Guid, host), &raw_response)

	switch apiErr.(type) {
	case nil:
		found = true
	case *errors.HttpNotFoundError:
		found = false
		apiErr = nil
	default:
		return
	}
	return
}

func (repo CloudControllerRouteRepository) CreateInSpace(host, path, domainGuid, spaceGuid string) (createdRoute models.Route, apiErr error) {
	data := fmt.Sprintf(`{"host":"%s","path":"%s","domain_guid":"%s","space_guid":"%s"}`, host, path, domainGuid, spaceGuid)

	resource := new(resources.RouteResource)
	apiErr = repo.gateway.CreateResource(repo.config.ApiEndpoint(), "/v2/routes?inline-relations-depth=1", strings.NewReader(data), resource)
	if apiErr != nil {
		return
	}

	createdRoute = resource.ToModel()
	return
}

func (repo CloudControllerRouteRepository) Bind(routeGuid, appGuid string) (apiErr error) {
	path := fmt.Sprintf("/v2/apps/%s/routes/%s", appGuid, routeGuid)
	return repo.gateway.UpdateResource(repo.config.ApiEndpoint(), path, nil)
}

func (repo CloudControllerRouteRepository) Unbind(routeGuid, appGuid string) (apiErr error) {
	path := fmt.Sprintf("/v2/apps/%s/routes/%s", appGuid, routeGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}

func (repo CloudControllerRouteRepository) Delete(routeGuid string) (apiErr error) {
	path := fmt.Sprintf("/v2/routes/%s", routeGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}
