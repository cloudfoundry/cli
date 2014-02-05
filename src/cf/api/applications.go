package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"encoding/json"
	"fmt"
	"generic"
	"net/url"
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
	Name               *string             `json:"name,omitempty"`
	Command            *string             `json:"command,omitempty"`
	State              *string             `json:"state,omitempty"`
	SpaceGuid          *string             `json:"space_guid,omitempty"`
	Instances          *int                `json:"instances,omitempty"`
	Memory             *uint64             `json:"memory,omitempty"`
	StackGuid          *string             `json:"stack_guid,omitempty"`
	Stack              *StackResource      `json:"stack,omitempty"`
	Routes             *[]AppRouteResource `json:"routes,omitempty"`
	Buildpack          *string             `json:"buildpack,omitempty"`
	EnvironmentJson    *map[string]string  `json:"environment_json,omitempty"`
	HealthCheckTimeout *int                `json:"health_check_timeout,omitempty"`
}

type ApplicationResource struct {
	Resource
	Entity ApplicationEntity
}

func NewApplicationEntityFromAppParams(app cf.AppParams) (entity ApplicationEntity) {
	if app.Has("buildpack") {
		buildpack := app.Get("buildpack").(string)
		entity.Buildpack = &buildpack
	}
	if app.Has("name") {
		name := app.Get("name").(string)
		entity.Name = &name
	}

	if app.Has("state") {
		state := strings.ToUpper(app.Get("state").(string))
		entity.State = &state
	}

	if app.Has("space_guid") {
		spaceGuid := app.Get("space_guid").(string)
		entity.SpaceGuid = &spaceGuid
	}

	if app.Has("instances") {
		instances := app.Get("instances").(int)
		entity.Instances = &instances
	}

	if app.Has("memory") {
		memory := app.Get("memory").(uint64)
		entity.Memory = &memory
	}

	if app.Has("stack_guid") {
		stackGuid := app.Get("stack_guid").(string)
		entity.StackGuid = &stackGuid
	}

	if app.Has("command") {
		command := app.Get("command").(string)
		entity.Command = &command
	}

	if app.Has("health_check_timeout") {
		healthCheckTimeout := app.Get("health_check_timeout").(int)
		entity.HealthCheckTimeout = &healthCheckTimeout
	}

	if app.Has("env") {
		envMap := app.Get("env").(generic.Map)
		if !envMap.IsEmpty() {
			environmentJson := map[string]string{}
			generic.Each(envMap, generic.Iterator(func(key, val interface{}) {
				environmentJson[key.(string)] = val.(string)
			}))
			entity.EnvironmentJson = &environmentJson
		}
	}

	return
}

func (resource ApplicationResource) ToFields() (app cf.ApplicationFields) {
	entity := resource.Entity
	app.Guid = resource.Metadata.Guid

	if entity.Name != nil {
		app.Name = *entity.Name
	}
	if entity.Memory != nil {
		app.Memory = uint64(*entity.Memory)
	}
	if entity.Instances != nil {
		app.InstanceCount = *entity.Instances
	}
	if entity.State != nil {
		app.State = strings.ToLower(*entity.State)
	}
	if entity.EnvironmentJson != nil {
		app.EnvironmentVars = *entity.EnvironmentJson
	}
	if entity.SpaceGuid != nil {
		app.SpaceGuid = *entity.SpaceGuid
	}
	return
}

func (resource ApplicationResource) ToModel() (app cf.Application) {
	app.ApplicationFields = resource.ToFields()

	entity := resource.Entity
	if entity.Stack != nil {
		app.Stack = entity.Stack.ToFields()
	}

	if entity.Routes != nil {
		for _, routeResource := range *entity.Routes {
			app.Routes = append(app.Routes, routeResource.ToModel())
		}
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
	data, err := repo.formatAppJSON(params)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Failed to marshal JSON", err)
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
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=%s&inline-relations-depth=1", repo.config.Target, repo.config.SpaceFields.Guid, url.QueryEscape("name:"+name))
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
	data, err := repo.formatAppJSON(params)
	if err != nil {
		apiResponse = net.NewApiResponseWithError("Failed to marshal JSON", err)
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

func (repo CloudControllerApplicationRepository) formatAppJSON(input cf.AppParams) (data string, err error) {
	appResource := NewApplicationEntityFromAppParams(input)
	bytes, err := json.Marshal(appResource)
	data = string(bytes)
	return
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, appGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
