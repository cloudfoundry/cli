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

//go:generate counterfeiter -o fakes/fake_route_repository.go . RouteRepository
type RouteRepository interface {
	ListRoutes(cb func(models.Route) bool) (apiErr error)
	ListAllRoutes(cb func(models.Route) bool) (apiErr error)
	Find(host string, domain models.DomainFields, path string) (route models.Route, apiErr error)
	Create(host string, domain models.DomainFields, path string) (createdRoute models.Route, apiErr error)
	CheckIfExists(host string, domain models.DomainFields, path string) (found bool, apiErr error)
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

func (repo CloudControllerRouteRepository) Find(host string, domain models.DomainFields, path string) (route models.Route, apiErr error) {
	if path != "" && !strings.HasPrefix(path, `/`) {
		path = `/` + path
	}

	found := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=%s", url.QueryEscape("host:"+host+";domain_guid:"+domain.Guid+";path:"+path)),
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

func (repo CloudControllerRouteRepository) Create(host string, domain models.DomainFields, path string) (createdRoute models.Route, apiErr error) {
	return repo.CreateInSpace(host, path, domain.Guid, repo.config.SpaceFields().Guid)
}

func (repo CloudControllerRouteRepository) CheckIfExists(host string, domain models.DomainFields, path string) (bool, error) {
	var raw_response interface{}

	u, err := url.Parse(repo.config.ApiEndpoint())
	if err != nil {
		return false, err
	}

	u.Path = fmt.Sprintf("/v2/routes/reserved/domain/%s/host/%s", domain.Guid, host)
	if path != "" {
		q := u.Query()
		q.Set("path", path)
		u.RawQuery = q.Encode()
	}

	err = repo.gateway.GetResource(u.String(), &raw_response)
	if err != nil {
		if _, ok := err.(*errors.HttpNotFoundError); ok {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (repo CloudControllerRouteRepository) CreateInSpace(host, path, domainGuid, spaceGuid string) (createdRoute models.Route, apiErr error) {
	if path != "" && !strings.HasPrefix(path, `/`) {
		path = `/` + path
	}

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
