package pushaction

import (
	"fmt"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
)

type CommandLineSettings struct {
	BuildpackName      string
	Command            string
	CurrentDirectory   string
	DiskQuota          uint64
	DockerImage        string
	HealthCheckTimeout int
	HealthCheckType    string
	Instances          int
	Memory             uint64
	Name               string
	ProvidedAppPath    string
	StackName          string
}

func (settings CommandLineSettings) ApplicationPath() string {
	if settings.ProvidedAppPath != "" {
		return settings.absoluteProvidedAppPath()
	}
	return settings.CurrentDirectory
}

func (settings CommandLineSettings) OverrideManifestSettings(app manifest.Application) manifest.Application {
	if settings.BuildpackName != "" {
		app.BuildpackName = settings.BuildpackName
	}

	if settings.Command != "" {
		app.Command = settings.Command
	}

	if settings.DiskQuota != 0 {
		app.DiskQuota = settings.DiskQuota
	}

	if settings.DockerImage != "" {
		app.DockerImage = settings.DockerImage
	}

	if settings.HealthCheckTimeout != 0 {
		app.HealthCheckTimeout = settings.HealthCheckTimeout
	}

	if settings.HealthCheckType != "" {
		app.HealthCheckType = settings.HealthCheckType
	}

	if settings.Instances != 0 {
		app.Instances = settings.Instances
	}

	if settings.Memory != 0 {
		app.Memory = settings.Memory
	}

	if settings.Name != "" {
		app.Name = settings.Name
	}

	if settings.ProvidedAppPath != "" {
		app.Path = settings.absoluteProvidedAppPath()
	}
	if app.Path == "" {
		app.Path = settings.CurrentDirectory
	}

	if settings.StackName != "" {
		app.StackName = settings.StackName
	}

	return app
}

func (settings CommandLineSettings) String() string {
	return fmt.Sprintf(
		"App Name: '%s', Buildpack: '%s', Command: '%s', CurrentDirectory: '%s', Disk Quota: '%d', Docker Image: '%s', Health Check Timeout: '%d', Health Check Type: '%s', Instances: '%d', Memory: '%d', Provided App Path: '%s', Stack: '%s'",
		settings.Name,
		settings.BuildpackName,
		settings.Command,
		settings.CurrentDirectory,
		settings.DiskQuota,
		settings.DockerImage,
		settings.HealthCheckTimeout,
		settings.HealthCheckType,
		settings.Instances,
		settings.Memory,
		settings.ProvidedAppPath,
		settings.StackName,
	)
}

func (settings CommandLineSettings) absoluteProvidedAppPath() string {
	if !filepath.IsAbs(settings.ProvidedAppPath) {
		return filepath.Join(settings.CurrentDirectory, settings.ProvidedAppPath)
	}
	return settings.ProvidedAppPath
}
