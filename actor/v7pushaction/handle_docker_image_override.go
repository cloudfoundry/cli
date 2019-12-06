package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func HandleDockerImageOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.DockerImage != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()

		if app.HasBuildpacks() {
			return manifest, translatableerror.ArgumentManifestMismatchError{
				Arg:              "--docker-image, -o",
				ManifestProperty: "buildpacks",
			}
		}

		if app.Path != "" {
			return manifest, translatableerror.ArgumentManifestMismatchError{
				Arg:              "--docker-image, -o",
				ManifestProperty: "path",
			}
		}

		if app.Docker == nil {
			emptyDockerInfo := manifestparser.Docker{}
			app.Docker = &emptyDockerInfo
		}

		app.Docker.Image = overrides.DockerImage
	}

	return manifest, nil
}
