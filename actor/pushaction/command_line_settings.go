package pushaction

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/v7/types"
	"code.cloudfoundry.org/cli/v7/util/manifest"
)

type CommandLineSettings struct {
	Buildpacks           []string
	Command              types.FilteredString
	CurrentDirectory     string
	DefaultRouteDomain   string
	DefaultRouteHostname string
	DiskQuota            uint64
	DockerImage          string
	DockerPassword       string
	DockerUsername       string
	DropletPath          string
	HealthCheckTimeout   uint64
	HealthCheckType      string
	Instances            types.NullInt
	Memory               uint64
	Name                 string
	NoHostname           bool
	NoRoute              bool
	ProvidedAppPath      string
	RandomRoute          bool
	RoutePath            string
	StackName            string
}

func (settings CommandLineSettings) OverrideManifestSettings(app manifest.Application) manifest.Application {
	if len(settings.Buildpacks) > 0 {
		app = settings.setBuildpacks(app)
	}

	if settings.Command.IsSet {
		app.Command = settings.Command
	}

	if settings.DefaultRouteDomain != "" {
		app.Domain = settings.DefaultRouteDomain
	}

	if settings.DefaultRouteHostname != "" {
		app.Hostname = settings.DefaultRouteHostname
		app.RandomRoute = false
	}

	if settings.DiskQuota != 0 {
		app.DiskQuota.ParseUint64Value(&settings.DiskQuota)
	}

	if settings.DockerImage != "" {
		app.DockerImage = settings.DockerImage
	}

	if settings.DockerUsername != "" {
		app.DockerUsername = settings.DockerUsername
	}

	if settings.DockerPassword != "" {
		app.DockerPassword = settings.DockerPassword
	}

	if settings.DropletPath != "" {
		app.DropletPath = settings.DropletPath
	}

	if settings.HealthCheckTimeout != 0 {
		app.HealthCheckTimeout = settings.HealthCheckTimeout
	}

	if settings.HealthCheckType != "" {
		app.HealthCheckType = settings.HealthCheckType
	}

	if settings.Instances.IsSet {
		app.Instances = settings.Instances
	}

	if settings.Memory != 0 {
		app.Memory.ParseUint64Value(&settings.Memory)
	}

	if settings.Name != "" {
		app.Name = settings.Name
	}

	if settings.NoHostname {
		app.NoHostname = true
	}

	if settings.NoRoute {
		app.NoRoute = true
	}

	if settings.ProvidedAppPath != "" {
		app.Path = settings.ProvidedAppPath
	}

	if app.Path == "" && app.DockerImage == "" && app.DropletPath == "" {
		app.Path = settings.CurrentDirectory
	}

	if settings.RandomRoute {
		app.RandomRoute = true
	}

	if settings.RoutePath != "" {
		app.RoutePath = settings.RoutePath
	}

	if settings.StackName != "" {
		app.StackName = settings.StackName
	}

	return app
}

func (settings CommandLineSettings) setBuildpacks(app manifest.Application) manifest.Application {
	app.Buildpack = types.FilteredString{}
	app.Buildpacks = nil

	if len(settings.Buildpacks) == 1 {
		app.Buildpack.ParseValue(settings.Buildpacks[0])
		return app
	}

	for _, bp := range settings.Buildpacks {
		app.Buildpacks = append(app.Buildpacks, bp)
	}

	return app
}

func (settings CommandLineSettings) String() string {
	return fmt.Sprintf(
		"App Name: '%s', Buildpacks: %s, Command: (%t, '%s'), CurrentDirectory: '%s', Disk Quota: '%d', Docker Image: '%s', Droplet: '%s', Health Check Timeout: '%d', Health Check Type: '%s', Instances: (%t, '%d'), Memory: '%d', Provided App Path: '%s', Stack: '%s', RoutePath: '%s', Domain: '%s', Hostname: '%s'",
		settings.Name,
		strings.Join(settings.Buildpacks, ", "),
		settings.Command.IsSet,
		settings.Command.Value,
		settings.CurrentDirectory,
		settings.DiskQuota,
		settings.DockerImage,
		settings.DropletPath,
		settings.HealthCheckTimeout,
		settings.HealthCheckType,
		settings.Instances.IsSet,
		settings.Instances.Value,
		settings.Memory,
		settings.ProvidedAppPath,
		settings.StackName,
		settings.RoutePath,
		settings.DefaultRouteDomain,
		settings.DefaultRouteHostname,
	)
}
