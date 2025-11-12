package v7pushaction

import (
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/util/manifestparser"
)

func HandleNoRouteOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.NoRoute {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()
		app.NoRoute = true
	}

	return manifest, nil
}
