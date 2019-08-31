package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func HandleBuildpacksOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
	if len(overrides.Buildpacks) > 0 {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		if manifest.Applications[0].Docker != nil {
			return manifest, translatableerror.ArgumentManifestMismatchError{
				Arg:              "--buildpack, -b",
				ManifestProperty: "docker",
			}
		}
		app := manifest.GetFirstApp()

		if overrides.Buildpacks[0] == "null" || overrides.Buildpacks[0] == "default" {
			app.SetBuildpacks([]string{})
		} else {
			app.SetBuildpacks(overrides.Buildpacks)
		}
	}

	return manifest, nil
}
