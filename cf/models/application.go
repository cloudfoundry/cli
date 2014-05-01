package models

import (
	"reflect"
	"strings"
)

type Application struct {
	ApplicationFields
	Stack  *Stack
	Routes []RouteSummary
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
	Guid             string
	Name             string
	BuildpackUrl     string
	Command          string
	DiskQuota        uint64 // in Megabytes
	EnvironmentVars  map[string]string
	InstanceCount    int
	Memory           uint64 // in Megabytes
	RunningInstances int
	State            string
	SpaceGuid        string
}

type AppParams struct {
	BuildpackUrl       *string
	Command            *string
	DiskQuota          *uint64
	Domain             *string
	EnvironmentVars    *map[string]string
	Guid               *string
	HealthCheckTimeout *int
	Host               *string
	InstanceCount      *int
	Memory             *uint64
	Name               *string
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
	if other.Domain != nil {
		app.Domain = other.Domain
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
	if other.Host != nil {
		app.Host = other.Host
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
	app.UseRandomHostname = app.UseRandomHostname || other.UseRandomHostname
}

func (app *AppParams) IsEmpty() bool {
	return reflect.DeepEqual(*app, AppParams{})
}
