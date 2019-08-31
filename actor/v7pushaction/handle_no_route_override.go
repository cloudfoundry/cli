package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func HandleNoRouteOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
	if overrides.NoRoute {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()
		app.NoRoute = overrides.NoRoute
	}

	return manifest, nil
}
