package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"encoding/json"
	"fmt"
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

func (resource AppRouteResource) ToFields() (route models.RouteFields) {
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host
	return
}

func (resource AppRouteResource) ToModel() (route models.RouteSummary) {
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
	DiskQuota          *uint64             `json:"disk_quota,omitempty"`
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

func NewApplicationEntityFromAppParams(app models.AppParams) ApplicationEntity {
	entity := ApplicationEntity{
		Buildpack:          app.BuildpackUrl,
		Name:               app.Name,
		SpaceGuid:          app.SpaceGuid,
		Instances:          app.InstanceCount,
		Memory:             app.Memory,
		DiskQuota:          app.DiskQuota,
		StackGuid:          app.StackGuid,
		Command:            app.Command,
		HealthCheckTimeout: app.HealthCheckTimeout,
	}
	if app.State != nil {
		state := strings.ToUpper(*app.State)
		entity.State = &state
	}
	if app.EnvironmentVars != nil && len(*app.EnvironmentVars) > 0 {
		entity.EnvironmentJson = app.EnvironmentVars
	}
	return entity
}

func (resource ApplicationResource) ToFields() (app models.ApplicationFields) {
	entity := resource.Entity
	app.Guid = resource.Metadata.Guid

	if entity.Name != nil {
		app.Name = *entity.Name
	}
	if entity.Memory != nil {
		app.Memory = *entity.Memory
	}
	if entity.DiskQuota != nil {
		app.DiskQuota = *entity.DiskQuota
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

func (resource ApplicationResource) ToModel() (app models.Application) {
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
	Create(params models.AppParams) (createdApp models.Application, apiResponse errors.Error)
	Read(name string) (app models.Application, apiResponse errors.Error)
	Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiResponse errors.Error)
	Delete(appGuid string) (apiResponse errors.Error)
}

type CloudControllerApplicationRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerApplicationRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerApplicationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationRepository) Create(params models.AppParams) (createdApp models.Application, apiResponse errors.Error) {
	data, err := repo.formatAppJSON(params)
	if err != nil {
		apiResponse = errors.NewErrorWithError("Failed to marshal JSON", err)
		return
	}

	path := fmt.Sprintf("%s/v2/apps", repo.config.ApiEndpoint())
	resource := new(ApplicationResource)
	apiResponse = repo.gateway.CreateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)
	if apiResponse != nil {
		return
	}

	createdApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Read(name string) (app models.Application, apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=%s&inline-relations-depth=1", repo.config.ApiEndpoint(), repo.config.SpaceFields().Guid, url.QueryEscape("name:"+name))
	appResources := new(PaginatedApplicationResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken(), appResources)
	if apiResponse != nil {
		return
	}

	if len(appResources.Resources) == 0 {
		apiResponse = errors.NewNotFoundError("%s %s not found", "App", name)
		return
	}

	res := appResources.Resources[0]
	app = res.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiResponse errors.Error) {
	data, err := repo.formatAppJSON(params)
	if err != nil {
		apiResponse = errors.NewErrorWithError("Failed to marshal JSON", err)
		return
	}

	path := fmt.Sprintf("%s/v2/apps/%s?inline-relations-depth=1", repo.config.ApiEndpoint(), appGuid)
	resource := new(ApplicationResource)
	apiResponse = repo.gateway.UpdateResourceForResponse(path, repo.config.AccessToken(), strings.NewReader(data), resource)
	if apiResponse != nil {
		return
	}

	updatedApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) formatAppJSON(input models.AppParams) (data string, err error) {
	appResource := NewApplicationEntityFromAppParams(input)
	bytes, err := json.Marshal(appResource)
	data = string(bytes)
	return
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.ApiEndpoint(), appGuid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken())
}
