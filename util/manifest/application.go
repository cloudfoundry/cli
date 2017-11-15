package manifest

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/types"
)

type Application struct {
	Buildpack      types.FilteredString
	Command        types.FilteredString
	DiskQuota      types.NullByteSizeInMb
	DockerImage    string
	DockerPassword string
	DockerUsername string
	Domain         string
	// EnvironmentVariables can be any valid json type (ie, strings not
	// guaranteed, although CLI only ships strings).
	EnvironmentVariables    map[string]string
	HealthCheckHTTPEndpoint string
	HealthCheckTimeout      int
	// HealthCheckType attribute defines the number of seconds that is allocated
	// for starting an application.
	HealthCheckType string
	Hostname        string
	Instances       types.NullInt
	Memory          types.NullByteSizeInMb
	Name            string
	NoHostname      bool
	NoRoute         bool
	Path            string
	RandomRoute     bool
	Routes          []string
	RoutePath       string
	Services        []string
	StackName       string

	DeprecatedDomain     interface{}
	DeprecatedDomains    interface{}
	DeprecatedHost       interface{}
	DeprecatedHosts      interface{}
	DeprecatedNoHostname interface{}
}

func (app Application) String() string {
	return fmt.Sprintf(
		"App Name: '%s', Buildpack IsSet: %t, Buildpack: '%s', Command IsSet: %t, Command: '%s', Disk Quota: '%s', Docker Image: '%s', Health Check HTTP Endpoint: '%s', Health Check Timeout: '%d', Health Check Type: '%s', Hostname: '%s', Instances IsSet: %t, Instances: '%d', Memory: '%s', No-Hostname: %t, No-Route: %t, Path: '%s', RandomRoute: %t, RoutePath: '%s', Routes: [%s], Services: [%s], Stack Name: '%s'",
		app.Name,
		app.Buildpack.IsSet,
		app.Buildpack.Value,
		app.Command.IsSet,
		app.Command.Value,
		app.DiskQuota,
		app.DockerImage,
		app.HealthCheckHTTPEndpoint,
		app.HealthCheckTimeout,
		app.HealthCheckType,
		app.Hostname,
		app.Instances.IsSet,
		app.Instances.Value,
		app.Memory,
		app.NoHostname,
		app.NoRoute,
		app.Path,
		app.RandomRoute,
		app.RoutePath,
		strings.Join(app.Routes, ", "),
		strings.Join(app.Services, ", "),
		app.StackName,
	)
}

func (app Application) MarshalYAML() (interface{}, error) {
	var m = rawManifestApplication{
		Buildpack:               app.Buildpack.Value,
		Command:                 app.Command.Value,
		Docker:                  rawDockerInfo{Image: app.DockerImage, Username: app.DockerUsername},
		EnvironmentVariables:    app.EnvironmentVariables,
		HealthCheckHTTPEndpoint: app.HealthCheckHTTPEndpoint,
		HealthCheckType:         app.HealthCheckType,
		Name:                    app.Name,
		NoRoute:                 app.NoRoute,
		Path:                    app.Path,
		Services:                app.Services,
		StackName:               app.StackName,
		Timeout:                 app.HealthCheckTimeout,
	}
	m.DiskQuota = app.DiskQuota.String()
	m.Memory = app.Memory.String()

	if app.Instances.IsSet {
		m.Instances = &app.Instances.Value
	}

	for _, route := range app.Routes {
		m.Routes = append(m.Routes, rawManifestRoute{Route: route})
	}

	return m, nil
}

func (app *Application) UnmarshalYAML(unmarshaller func(interface{}) error) error {
	var m rawManifestApplication

	err := unmarshaller(&m)
	if err != nil {
		return err
	}
	app.DeprecatedDomain = m.DeprecatedDomain
	app.DeprecatedDomains = m.DeprecatedDomains
	app.DeprecatedHost = m.DeprecatedHost
	app.DeprecatedHosts = m.DeprecatedHosts
	app.DeprecatedNoHostname = m.DeprecatedNoHostname
	app.DockerImage = m.Docker.Image
	app.DockerUsername = m.Docker.Username
	app.HealthCheckHTTPEndpoint = m.HealthCheckHTTPEndpoint
	app.HealthCheckType = m.HealthCheckType
	app.Name = m.Name
	app.NoRoute = m.NoRoute
	app.Path = m.Path
	app.RandomRoute = m.RandomRoute
	app.Services = m.Services
	app.StackName = m.StackName
	app.HealthCheckTimeout = m.Timeout
	app.EnvironmentVariables = m.EnvironmentVariables

	app.Instances.ParseIntValue(m.Instances)

	if fmtErr := app.DiskQuota.ParseStringValue(m.DiskQuota); fmtErr != nil {
		return fmtErr
	}

	if fmtErr := app.Memory.ParseStringValue(m.Memory); fmtErr != nil {
		return fmtErr
	}

	for _, route := range m.Routes {
		app.Routes = append(app.Routes, route.Route)
	}

	// "null" values are identical to non-existant values in YAML. In order to
	// detect if an explicit null is given, a manual existance check is required.
	exists := map[string]interface{}{}
	err = unmarshaller(&exists)
	if err != nil {
		return err
	}

	if _, ok := exists["buildpack"]; ok {
		app.Buildpack.ParseValue(m.Buildpack)
		app.Buildpack.IsSet = true
	}

	if _, ok := exists["command"]; ok {
		app.Command.ParseValue(m.Command)
		app.Command.IsSet = true
	}

	return nil
}
