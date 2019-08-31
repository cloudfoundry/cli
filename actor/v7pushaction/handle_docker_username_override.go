package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func HandleDockerUsernameOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
	if overrides.DockerUsername != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()

		if app.Docker == nil {
			emptyDockerInfo := pushmanifestparser.Docker{}
			app.Docker = &emptyDockerInfo
		}

		app.Docker.Username = overrides.DockerUsername
	}

	return manifest, nil
}
