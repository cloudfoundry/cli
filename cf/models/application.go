package models

import (
	"reflect"
	"strings"
	"time"
)

type Application struct {
	ApplicationFields
	Stack    *Stack
	Routes   []RouteSummary
	Services []ServicePlanSummary
}

func (model Application) HasRoute(route Route) bool {
	for _, boundRoute := range model.Routes {
		if boundRoute.Guid == route.Guid {
			return true
		}
	}
	return false
}

func (model Application) ToParams() (params AppParams) {
	state := strings.ToUpper(model.State)
	params = AppParams{
		Guid:            &model.Guid,
		Name:            &model.Name,
		BuildpackUrl:    &model.BuildpackUrl,
		Command:         &model.Command,
		DiskQuota:       &model.DiskQuota,
		InstanceCount:   &model.InstanceCount,
		Memory:          &model.Memory,
		State:           &state,
		SpaceGuid:       &model.SpaceGuid,
		EnvironmentVars: &model.EnvironmentVars,
	}

	if model.Stack != nil {
		params.StackGuid = &model.Stack.Guid
	}

	return
}

type ApplicationFields struct {
	Guid                 string
	Name                 string
	BuildpackUrl         string
	Command              string
	Diego                bool
	DetectedStartCommand string
	DiskQuota            int64 // in Megabytes
	EnvironmentVars      map[string]interface{}
	InstanceCount        int
	Memory               int64 // in Megabytes
	RunningInstances     int
	HealthCheckTimeout   int
	State                string
	SpaceGuid            string
	PackageUpdatedAt     *time.Time
	PackageState         string
	StagingFailedReason  string
}

type AppParams struct {
	BuildpackUrl       *string
	Command            *string
	DiskQuota          *int64
	Domains            *[]string
	EnvironmentVars    *map[string]interface{}
	Guid               *string
	HealthCheckTimeout *int
	Hosts              *[]string
	InstanceCount      *int
	Memory             *int64
	Name               *string
	NoHostname         bool
	NoRoute            bool
	UseRandomHostname  bool
	Path               *string
	ServicesToBind     *[]string
	SpaceGuid          *string
	StackGuid          *string
	StackName          *string
	State              *string
}

func (app *AppParams) Merge(other *AppParams) {
	if other.BuildpackUrl != nil {
		app.BuildpackUrl = other.BuildpackUrl
	}
	if other.Command != nil {
		app.Command = other.Command
	}
	if other.DiskQuota != nil {
		app.DiskQuota = other.DiskQuota
	}
	if other.Domains != nil {
		app.Domains = other.Domains
	}
	if other.EnvironmentVars != nil {
		app.EnvironmentVars = other.EnvironmentVars
	}
	if other.Guid != nil {
		app.Guid = other.Guid
	}
	if other.HealthCheckTimeout != nil {
		app.HealthCheckTimeout = other.HealthCheckTimeout
	}
	if other.Hosts != nil {
		app.Hosts = other.Hosts
	}
	if other.InstanceCount != nil {
		app.InstanceCount = other.InstanceCount
	}
	if other.DiskQuota != nil {
		app.DiskQuota = other.DiskQuota
	}
	if other.Memory != nil {
		app.Memory = other.Memory
	}
	if other.Name != nil {
		app.Name = other.Name
	}
	if other.Path != nil {
		app.Path = other.Path
	}
	if other.ServicesToBind != nil {
		app.ServicesToBind = other.ServicesToBind
	}
	if other.SpaceGuid != nil {
		app.SpaceGuid = other.SpaceGuid
	}
	if other.StackGuid != nil {
		app.StackGuid = other.StackGuid
	}
	if other.StackName != nil {
		app.StackName = other.StackName
	}
	if other.State != nil {
		app.State = other.State
	}

	app.NoRoute = app.NoRoute || other.NoRoute
	app.NoHostname = app.NoHostname || other.NoHostname
	app.UseRandomHostname = app.UseRandomHostname || other.UseRandomHostname
}

func (app *AppParams) IsEmpty() bool {
	return reflect.DeepEqual(*app, AppParams{})
}

func (app *AppParams) IsHostEmpty() bool {
	return app.Hosts == nil || len(*app.Hosts) == 0
}
