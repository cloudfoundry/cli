package api

import (
	"bytes"
	"encoding/json"
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
	CreateInSpace(host, path, domainGuid, spaceGuid string, port int, randomPort bool) (createdRoute models.Route, apiErr error)
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

func normalizedPath(path string) string {
	if path != "" && !strings.HasPrefix(path, `/`) {
		return `/` + path
	}

	return path
}

func queryStringForRouteSearch(host, guid, path string) string {
	args := []string{
		fmt.Sprintf("host:%s", host),
		fmt.Sprintf("domain_guid:%s", guid),
	}

	if path != "" {
		args = append(args, fmt.Sprintf("path:%s", normalizedPath(path)))
	}

	return url.QueryEscape(strings.Join(args, ";"))
}

func (repo CloudControllerRouteRepository) Find(host string, domain models.DomainFields, path string) (route models.Route, apiErr error) {
	queryString := queryStringForRouteSearch(host, domain.Guid, path)

	found := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=%s", queryString),
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
	var port int
	var randomRoute bool
	return repo.CreateInSpace(host, path, domain.Guid, repo.config.SpaceFields().Guid, port, randomRoute)
}

func (repo CloudControllerRouteRepository) CheckIfExists(host string, domain models.DomainFields, path string) (bool, error) {
	path = normalizedPath(path)

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

	var raw_response interface{}
	err = repo.gateway.GetResource(u.String(), &raw_response)
	if err != nil {
		if _, ok := err.(*errors.HttpNotFoundError); ok {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (repo CloudControllerRouteRepository) CreateInSpace(host, path, domainGuid, spaceGuid string, port int, randomPort bool) (models.Route, error) {
	path = normalizedPath(path)

	body := struct {
		Host         string `json:"host,omitempty"`
		Path         string `json:"path,omitempty"`
		Port         int    `json:"port,omitempty"`
		DomainGuid   string `json:"domain_guid"`
		SpaceGuid    string `json:"space_guid"`
		GeneratePort bool   `json:"generate_port"`
	}{host, path, port, domainGuid, spaceGuid, randomPort}

	data, err := json.Marshal(body)
	if err != nil {
		return models.Route{}, err
	}

	resource := new(resources.RouteResource)
	err = repo.gateway.CreateResource(
		repo.config.ApiEndpoint(),
		"/v2/routes?inline-relations-depth=1",
		bytes.NewReader(data),
		resource,
	)
	if err != nil {
		return models.Route{}, err
	}

	return resource.ToModel(), nil
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
