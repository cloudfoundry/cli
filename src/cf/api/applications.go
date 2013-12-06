package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"regexp"
	"strings"
)

type PaginatedApplicationResources struct {
	Resources []ApplicationResource
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

type ApplicationEntity struct {
	Name            string
	State           string
	Instances       int
	Memory          int
	Stack           StackResource
	Routes          []AppRouteResource
	EnvironmentJson map[string]string `json:"environment_json"`
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

type AppRouteEntity struct {
	Host   string
	Domain Resource
}

type ApplicationRepository interface {
	Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (createdApp cf.Application, apiResponse net.ApiResponse)
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

func (repo CloudControllerApplicationRepository) Create(name, buildpackUrl, stackGuid, command string, memory uint64, instances int) (createdApp cf.Application, apiResponse net.ApiResponse) {
	params := cf.NewAppParams()
	params.Fields["name"] = name
	params.Fields["buildpack"] = buildpackUrl
	params.Fields["command"] = command
	params.Fields["instances"] = instances
	params.Fields["memory"] = memory
	params.Fields["stack_guid"] = stackGuid

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

func (repo CloudControllerApplicationRepository) formatAppJSON(params cf.AppParams) (data string, apiResponse net.ApiResponse) {
	delete(params.Fields, "guid")

	command, ok := params.Fields["command"]
	if ok && command.(string) == "null" {
		params.Fields["command"] = ""
	} else if ok {
		params.Fields["command"] = stringOrNull(command.(string))
	}

	buildpack, ok := params.Fields["buildpack"]
	if ok {
		params.Fields["buildpack"] = stringOrNull(buildpack.(string))
	}

	stackGuid, ok := params.Fields["stack_guid"]
	if ok {
		params.Fields["stack_guid"] = stringOrNull(stackGuid.(string))
	}

	name, ok := params.Fields["name"]
	if ok {
		reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
		if !reg.MatchString(name.(string)) {
			apiResponse = net.NewApiResponseWithMessage("App name is invalid: name can only contain letters, numbers, underscores and hyphens")
			return
		}
	}

	vals := []string{}

	if !params.Fields.IsEmpty() {
		vals = append(vals, mapToJsonValues(params.Fields)...)
	}

	if !params.EnvironmentVars.IsEmpty() {
		envVal := fmt.Sprintf(`"environment_json":{%s}`, strings.Join(mapToJsonValues(params.EnvironmentVars), ","))
		vals = append(vals, envVal)
	}

	data = fmt.Sprintf("{%s}", strings.Join(vals, ","))
	return
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, appGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
