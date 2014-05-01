package resources

import (
	"github.com/cloudfoundry/cli/cf/models"
	"strings"
)

type PaginatedApplicationResources struct {
	Resources []ApplicationResource
}

type AppRouteEntity struct {
	Host   string
	Domain struct {
		Resource
		Entity struct {
			Name string
		}
	}
}

type AppRouteResource struct {
	Resource
	Entity AppRouteEntity
}

type AppFileResource struct {
	Path string `json:"fn"`
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
}

type ApplicationResource struct {
	Resource
	Entity ApplicationEntity
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

func (resource AppRouteResource) ToFields() (route models.RouteSummary) {
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host
	return
}

func (resource AppRouteResource) ToModel() (route models.RouteSummary) {
	route.Guid = resource.Metadata.Guid
	route.Host = resource.Entity.Host
	route.Domain.Guid = resource.Entity.Domain.Metadata.Guid
	route.Domain.Name = resource.Entity.Domain.Entity.Name
	return
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
	if app.EnvironmentVars != nil && *app.EnvironmentVars != nil {
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
