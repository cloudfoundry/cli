package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func HandleDropletPathOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.DropletPath != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()

		if app.Docker != nil {
			return manifest, translatableerror.ArgumentManifestMismatchError{
				Arg:              "--droplet",
				ManifestProperty: "docker",
			}
		}

		if app.Path != "" {
			return manifest, translatableerror.ArgumentManifestMismatchError{
				Arg:              "--droplet",
				ManifestProperty: "path",
			}
		}

		if app.HasBuildpacks() {
			return manifest, translatableerror.ArgumentManifestMismatchError{
				Arg:              "--droplet",
				ManifestProperty: "buildpacks",
			}
		}
	}

	return manifest, nil
}
