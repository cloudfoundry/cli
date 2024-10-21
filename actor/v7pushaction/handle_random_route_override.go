package v7pushaction

import (
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/util/manifestparser"
)

func HandleRandomRouteOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.RandomRoute {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()
		app.RandomRoute = overrides.RandomRoute
	}

	return manifest, nil
}
