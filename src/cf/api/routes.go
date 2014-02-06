package api

import (
	"cf"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type PaginatedRouteResources struct {
	Resources []RouteResource `json:"resources"`
	NextUrl   string          `json:"next_url"`
}

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
	ListRoutes(stop chan bool) (routesChan chan []models.Route, statusChan chan net.ApiResponse)
	FindByHost(host string) (route models.Route, apiResponse net.ApiResponse)
	FindByHostAndDomain(host, domain string) (route models.Route, apiResponse net.ApiResponse)
	Create(host, domainGuid string) (createdRoute models.Route, apiResponse net.ApiResponse)
	CreateInSpace(host, domainGuid, spaceGuid string) (createdRoute models.Route, apiResponse net.ApiResponse)
	Bind(routeGuid, appGuid string) (apiResponse net.ApiResponse)
	Unbind(routeGuid, appGuid string) (apiResponse net.ApiResponse)
	Delete(routeGuid string) (apiResponse net.ApiResponse)
}

type CloudControllerRouteRepository struct {
	config     *configuration.Configuration
	gateway    net.Gateway
	domainRepo DomainRepository
}

func NewCloudControllerRouteRepository(config *configuration.Configuration, gateway net.Gateway, domainRepo DomainRepository) (repo CloudControllerRouteRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.domainRepo = domainRepo
	return
}

func (repo CloudControllerRouteRepository) ListRoutes(stop chan bool) (routesChan chan []models.Route, statusChan chan net.ApiResponse) {
	routesChan = make(chan []models.Route, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		path := fmt.Sprintf("/v2/routes?inline-relations-depth=1")

	loop:
		for path != "" {
			select {
			case <-stop:
				break loop
			default:
				var (
					routes      []models.Route
					apiResponse net.ApiResponse
				)
				routes, path, apiResponse = repo.findNextWithPath(path)
				if apiResponse.IsNotSuccessful() {
					statusChan <- apiResponse
					close(routesChan)
					close(statusChan)
					return
				}

				if len(routes) > 0 {
					routesChan <- routes
				}
			}
		}
		close(routesChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerRouteRepository) FindByHost(host string) (route models.Route, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=%s", url.QueryEscape("host:"+host))
	return repo.findOneWithPath(path)
}

func (repo CloudControllerRouteRepository) FindByHostAndDomain(host, domainName string) (route models.Route, apiResponse net.ApiResponse) {
	domain, apiResponse := repo.domainRepo.FindByName(domainName)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("/v2/routes?inline-relations-depth=1&q=%s", url.QueryEscape("host:"+host+";domain_guid:"+domain.Guid))
	route, apiResponse = repo.findOneWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	route.Domain = domain.DomainFields
	return
}

func (repo CloudControllerRouteRepository) findOneWithPath(path string) (route models.Route, apiResponse net.ApiResponse) {
	routes, _, apiResponse := repo.findNextWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(routes) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Route not found")
		return
	}

	route = routes[0]
	return
}

func (repo CloudControllerRouteRepository) findNextWithPath(path string) (routes []models.Route, nextUrl string, apiResponse net.ApiResponse) {
	routesResources := new(PaginatedRouteResources)
	apiResponse = repo.gateway.GetResource(repo.config.Target+path, repo.config.AccessToken, routesResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = routesResources.NextUrl

	for _, routeResponse := range routesResources.Resources {
		routes = append(routes, routeResponse.ToModel())
	}
	return
}

func (repo CloudControllerRouteRepository) Create(host, domainGuid string) (createdRoute models.Route, apiResponse net.ApiResponse) {
	return repo.CreateInSpace(host, domainGuid, repo.config.SpaceFields.Guid)
}

func (repo CloudControllerRouteRepository) CreateInSpace(host, domainGuid, spaceGuid string) (createdRoute models.Route, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/routes?inline-relations-depth=1", repo.config.Target)
	data := fmt.Sprintf(`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`, host, domainGuid, spaceGuid)

	resource := new(RouteResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdRoute = resource.ToModel()
	return
}

func (repo CloudControllerRouteRepository) Bind(routeGuid, appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, appGuid, routeGuid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, nil)
}

func (repo CloudControllerRouteRepository) Unbind(routeGuid, appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s/routes/%s", repo.config.Target, appGuid, routeGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerRouteRepository) Delete(routeGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/routes/%s", repo.config.Target, routeGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
