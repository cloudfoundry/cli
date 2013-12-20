package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"generic"
	"regexp"
	"strings"
)

type AppRouteEntity struct {
	Host   string
	Domain Resource
}

type AppRouteResource struct {
	Resource
	Entity AppRouteEntity
}

func (resource AppRouteResource) ToFields() (route cf.RouteFields) {
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host
	return
}

func (resource AppRouteResource) ToModel() (route cf.RouteSummary) {
	route.RouteFields = resource.ToFields()
	route.Domain.Guid = resource.Entity.Domain.Metadata.Guid
	route.Domain.Name = resource.Entity.Domain.Entity.Name
	return
}

type ApplicationEntity struct {
	Name            string
	State           string
	SpaceGuid       string `json:"space_guid"`
	Instances       int
	Memory          int
	Stack           StackResource
	Routes          []AppRouteResource
	EnvironmentJson map[string]string `json:"environment_json"`
}

type ApplicationResource struct {
	Resource
	Entity ApplicationEntity
}

func (resource ApplicationResource) ToFields() (app cf.ApplicationFields) {
	app.Guid = resource.Metadata.Guid
	app.Name = resource.Entity.Name
	app.EnvironmentVars = resource.Entity.EnvironmentJson
	app.State = strings.ToLower(resource.Entity.State)
	app.InstanceCount = resource.Entity.Instances
	app.Memory = uint64(resource.Entity.Memory)
	app.SpaceGuid = resource.Entity.SpaceGuid
	return
}

func (resource ApplicationResource) ToModel() (app cf.Application) {
	app.ApplicationFields = resource.ToFields()
	app.Stack = resource.Entity.Stack.ToFields()

	for _, routeResource := range resource.Entity.Routes {
		app.Routes = append(app.Routes, routeResource.ToModel())
	}
	return
}

type PaginatedApplicationResources struct {
	Resources []ApplicationResource
}

type ApplicationRepository interface {
	Create(params cf.AppParams) (createdApp cf.Application, apiResponse net.ApiResponse)
	Read(name string) (app cf.Application, apiResponse net.ApiResponse)
	Update(appGuid string, params cf.AppParams) (updatedApp cf.Application, apiResponse net.ApiResponse)
	Delete(appGuid string) (apiResponse net.ApiResponse)
}

type CloudControllerApplicationRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerApplicationRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerApplicationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationRepository) Create(params cf.AppParams) (createdApp cf.Application, apiResponse net.ApiResponse) {
	data, apiResponse := repo.formatAppJSON(params)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/apps", repo.config.Target)
	resource := new(ApplicationResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	createdApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Read(name string) (app cf.Application, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.SpaceFields.Guid, "%3A"+name)
	appResources := new(PaginatedApplicationResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, appResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(appResources.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "App", name)
		return
	}

	res := appResources.Resources[0]
	app = res.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Update(appGuid string, params cf.AppParams) (updatedApp cf.Application, apiResponse net.ApiResponse) {
	data, apiResponse := repo.formatAppJSON(params)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/v2/apps/%s?inline-relations-depth=1", repo.config.Target, appGuid)
	resource := new(ApplicationResource)
	apiResponse = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken, strings.NewReader(data), resource)
	if apiResponse.IsNotSuccessful() {
		return
	}

	updatedApp = resource.ToModel()
	return
}

var allowedAppKeys = []string{
	"buildpack",
	"command",
	"instances",
	"memory",
	"name",
	"space_guid",
	"stack_guid",
	"state",
	"host",
	"domain",
}

func (repo CloudControllerApplicationRepository) formatAppJSON(input cf.AppParams) (data string, apiResponse net.ApiResponse) {
	params := generic.NewEmptyMap()
	for _, allowedKey := range allowedAppKeys {
		if input.Has(allowedKey) {
			params.Set(allowedKey, input.Get(allowedKey))
		}
	}

	if params.Has("command") && params.Get("command").(string) == "null" {
		params.Set("command", "")
	} else if params.Has("command") {
		params.Set("command", stringOrNull(params.Get("command")))
	}

	if params.Has("buildpack") {
		params.Set("buildpack", stringOrNull(params.Get("buildpack")))
	}

	if params.Has("stack_guid") {
		params.Set("stack_guid", stringOrNull(params.Get("stack_guid")))
	}

	if params.Has("state") {
		params.Set("state", strings.ToUpper(params.Get("state").(string)))
	}

	if params.Has("name") {
		reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
		if !reg.MatchString(params.Get("name").(string)) {
			apiResponse = net.NewApiResponseWithMessage("App name is invalid: name can only contain letters, numbers, underscores and hyphens")
			return
		}
	}

	vals := []string{}

	if !params.IsEmpty() {
		vals = append(vals, mapToJsonValues(params)...)
	}
	if input.Has("env") {
		envVars := input.Get("env").(generic.Map)
		if !envVars.IsEmpty() {
			envVal := fmt.Sprintf(`"environment_json":{%s}`, strings.Join(mapToJsonValues(envVars), ","))
			vals = append(vals, envVal)
		}

	}
	data = fmt.Sprintf("{%s}", strings.Join(vals, ","))
	return
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, appGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
