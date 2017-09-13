package models

import (
	"os"
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
		if boundRoute.GUID == route.GUID {
			return true
		}
	}
	return false
}

func (model Application) ToParams() AppParams {
	state := strings.ToUpper(model.State)
	params := AppParams{
		GUID:                    &model.GUID,
		Name:                    &model.Name,
		BuildpackURL:            &model.BuildpackURL,
		Command:                 &model.Command,
		DiskQuota:               &model.DiskQuota,
		InstanceCount:           &model.InstanceCount,
		HealthCheckType:         &model.HealthCheckType,
		HealthCheckHTTPEndpoint: &model.HealthCheckHTTPEndpoint,
		Memory:                  &model.Memory,
		State:                   &state,
		SpaceGUID:               &model.SpaceGUID,
		EnvironmentVars:         &model.EnvironmentVars,
		DockerImage:             &model.DockerImage,
	}

	if model.Stack != nil {
		params.StackGUID = &model.Stack.GUID
	}

	return params
}

type ApplicationFields struct {
	GUID                    string
	Name                    string
	BuildpackURL            string
	Command                 string
	Diego                   bool
	DetectedStartCommand    string
	DiskQuota               int64 // in Megabytes
	EnvironmentVars         map[string]interface{}
	InstanceCount           int
	Memory                  int64 // in Megabytes
	RunningInstances        int
	HealthCheckType         string
	HealthCheckHTTPEndpoint string
	HealthCheckTimeout      int
	State                   string
	SpaceGUID               string
	StackGUID               string
	PackageUpdatedAt        *time.Time
	PackageState            string
	StagingFailedReason     string
	Buildpack               string
	DetectedBuildpack       string
	DockerImage             string
	EnableSSH               bool
	AppPorts                []int
}

const (
	ApplicationStateStopped  = "stopped"
	ApplicationStateStarted  = "started"
	ApplicationStateRunning  = "running"
	ApplicationStateCrashed  = "crashed"
	ApplicationStateFlapping = "flapping"
	ApplicationStateDown     = "down"
	ApplicationStateStarting = "starting"
)

type AppParams struct {
	BuildpackURL            *string
	Command                 *string
	DiskQuota               *int64
	Domains                 []string
	EnvironmentVars         *map[string]interface{}
	GUID                    *string
	HealthCheckType         *string
	HealthCheckHTTPEndpoint *string
	HealthCheckTimeout      *int
	DockerImage             *string
	DockerUsername          *string
	DockerPassword          *string
	Diego                   *bool
	EnableSSH               *bool
	Hosts                   []string
	RoutePath               *string
	InstanceCount           *int
	Memory                  *int64
	Name                    *string
	NoHostname              *bool
	NoRoute                 bool
	UseRandomRoute          bool
	UseRandomPort           bool
	Path                    *string
	ServicesToBind          []string
	SpaceGUID               *string
	StackGUID               *string
	StackName               *string
	State                   *string
	PackageUpdatedAt        *time.Time
	AppPorts                *[]int
	Routes                  []ManifestRoute
}

func (app *AppParams) Merge(flagContext *AppParams) {
	if flagContext.AppPorts != nil {
		app.AppPorts = flagContext.AppPorts
	}
	if flagContext.BuildpackURL != nil {
		app.BuildpackURL = flagContext.BuildpackURL
	}
	if flagContext.Command != nil {
		app.Command = flagContext.Command
	}
	if flagContext.DiskQuota != nil {
		app.DiskQuota = flagContext.DiskQuota
	}
	if flagContext.DockerImage != nil {
		app.DockerImage = flagContext.DockerImage
	}

	switch {
	case flagContext.DockerUsername != nil:
		app.DockerUsername = flagContext.DockerUsername
		// the password is always non-nil after we parse the flag context
		app.DockerPassword = flagContext.DockerPassword
	case app.DockerUsername != nil:
		password := os.Getenv("CF_DOCKER_PASSWORD")
		// if the password is empty, we will get a CC error
		app.DockerPassword = &password
	}

	if flagContext.Domains != nil {
		app.Domains = flagContext.Domains
	}
	if flagContext.EnableSSH != nil {
		app.EnableSSH = flagContext.EnableSSH
	}
	if flagContext.EnvironmentVars != nil {
		app.EnvironmentVars = flagContext.EnvironmentVars
	}
	if flagContext.GUID != nil {
		app.GUID = flagContext.GUID
	}
	if flagContext.HealthCheckType != nil {
		app.HealthCheckType = flagContext.HealthCheckType
	}
	if flagContext.HealthCheckHTTPEndpoint != nil {
		app.HealthCheckHTTPEndpoint = flagContext.HealthCheckHTTPEndpoint
	}
	if flagContext.HealthCheckTimeout != nil {
		app.HealthCheckTimeout = flagContext.HealthCheckTimeout
	}
	if flagContext.Hosts != nil {
		app.Hosts = flagContext.Hosts
	}
	if flagContext.InstanceCount != nil {
		app.InstanceCount = flagContext.InstanceCount
	}
	if flagContext.Memory != nil {
		app.Memory = flagContext.Memory
	}
	if flagContext.Name != nil {
		app.Name = flagContext.Name
	}
	if flagContext.Path != nil {
		app.Path = flagContext.Path
	}
	if flagContext.RoutePath != nil {
		app.RoutePath = flagContext.RoutePath
	}
	if flagContext.ServicesToBind != nil {
		app.ServicesToBind = flagContext.ServicesToBind
	}
	if flagContext.SpaceGUID != nil {
		app.SpaceGUID = flagContext.SpaceGUID
	}
	if flagContext.StackGUID != nil {
		app.StackGUID = flagContext.StackGUID
	}
	if flagContext.StackName != nil {
		app.StackName = flagContext.StackName
	}
	if flagContext.State != nil {
		app.State = flagContext.State
	}

	app.NoRoute = app.NoRoute || flagContext.NoRoute
	noHostBool := app.IsNoHostnameTrue() || flagContext.IsNoHostnameTrue()
	app.NoHostname = &noHostBool
	app.UseRandomRoute = app.UseRandomRoute || flagContext.UseRandomRoute
}

func (app *AppParams) IsEmpty() bool {
	noHostBool := false
	return reflect.DeepEqual(*app, AppParams{NoHostname: &noHostBool})
}

func (app *AppParams) IsHostEmpty() bool {
	return app.Hosts == nil || len(app.Hosts) == 0
}

func (app *AppParams) IsNoHostnameTrue() bool {
	if app.NoHostname == nil {
		return false
	}
	return *app.NoHostname
}
