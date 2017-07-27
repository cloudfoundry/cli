package pushaction

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
)

type CommandLineSettings struct {
	CurrentDirectory string
	DockerImage      string
	Name             string
	ProvidedAppPath  string
}

func (settings CommandLineSettings) ApplicationPath() string {
	if settings.ProvidedAppPath != "" {
		return settings.absoluteProvidedAppPath()
	}
	return settings.CurrentDirectory
}

func (settings CommandLineSettings) OverrideManifestSettings(app manifest.Application) manifest.Application {
	if settings.Name != "" {
		app.Name = settings.Name
	}

	if settings.ProvidedAppPath != "" {
		app.Path = settings.absoluteProvidedAppPath()
	}
	if app.Path == "" {
		app.Path = settings.CurrentDirectory
	}

	return app
}

func (settings CommandLineSettings) absoluteProvidedAppPath() string {
	if !filepath.IsAbs(settings.ProvidedAppPath) {
		return filepath.Join(settings.CurrentDirectory, settings.ProvidedAppPath)
	}
	return settings.ProvidedAppPath
}
