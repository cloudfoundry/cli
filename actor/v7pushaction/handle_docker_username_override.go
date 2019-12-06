package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func HandleDockerUsernameOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.DockerUsername != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()

		if app.Docker == nil {
			emptyDockerInfo := manifestparser.Docker{}
			app.Docker = &emptyDockerInfo
		}

		app.Docker.Username = overrides.DockerUsername
	}

	return manifest, nil
}
