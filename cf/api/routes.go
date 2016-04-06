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
	"github.com/google/go-querystring/query"
)

//go:generate counterfeiter . RouteRepository
type RouteRepository interface {
	ListRoutes(cb func(models.Route) bool) (apiErr error)
	ListAllRoutes(cb func(models.Route) bool) (apiErr error)
	Find(host string, domain models.DomainFields, path string, port int) (route models.Route, apiErr error)
	Create(host string, domain models.DomainFields, path string, useRandomPort bool) (createdRoute models.Route, apiErr error)
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

func queryStringForRouteSearch(host, guid, path string, port int) string {
	args := []string{
		fmt.Sprintf("host:%s", host),
		fmt.Sprintf("domain_guid:%s", guid),
	}

	if path != "" {
		args = append(args, fmt.Sprintf("path:%s", normalizedPath(path)))
	}

	if port != 0 {
		args = append(args, fmt.Sprintf("port:%d", port))
	}

	return strings.Join(args, ";")
}

func (repo CloudControllerRouteRepository) Find(host string, domain models.DomainFields, path string, port int) (models.Route, error) {
	var route models.Route
	queryString := queryStringForRouteSearch(host, domain.Guid, path, port)

	q := struct {
		Query                string `url:"q"`
		InlineRelationsDepth int    `url:"inline-relations-depth"`
	}{queryString, 1}

	opt, _ := query.Values(q)

	found := false
	apiErr := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/routes?%s", opt.Encode()),
		resources.RouteResource{},
		func(resource interface{}) bool {
			keepSearching := true
			route = resource.(resources.RouteResource).ToModel()
			if doesNotMatchVersionSpecificAttributes(route, path, port) {
				return keepSearching
			}

			found = true
			return !keepSearching
		})

	if apiErr == nil && !found {
		apiErr = errors.NewModelNotFoundError("Route", host)
	}

	return route, apiErr
}

func doesNotMatchVersionSpecificAttributes(route models.Route, path string, port int) bool {
	return normalizedPath(route.Path) != normalizedPath(path) || route.Port != port
}

func (repo CloudControllerRouteRepository) Create(host string, domain models.DomainFields, path string, useRandomPort bool) (createdRoute models.Route, apiErr error) {
	var port int
	return repo.CreateInSpace(host, path, domain.Guid, repo.config.SpaceFields().Guid, port, useRandomPort)
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
		Host       string `json:"host,omitempty"`
		Path       string `json:"path,omitempty"`
		Port       int    `json:"port,omitempty"`
		DomainGuid string `json:"domain_guid"`
		SpaceGuid  string `json:"space_guid"`
	}{host, path, port, domainGuid, spaceGuid}

	data, err := json.Marshal(body)
	if err != nil {
		return models.Route{}, err
	}

	q := struct {
		GeneratePort         bool `url:"generate_port,omitempty"`
		InlineRelationsDepth int  `url:"inline-relations-depth"`
	}{randomPort, 1}

	opt, _ := query.Values(q)
	uriFragment := "/v2/routes?" + opt.Encode()

	resource := new(resources.RouteResource)
	err = repo.gateway.CreateResource(
		repo.config.ApiEndpoint(),
		uriFragment,
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
