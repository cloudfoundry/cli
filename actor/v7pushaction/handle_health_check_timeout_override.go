package v7pushaction

import (
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/util/manifestparser"
)

func HandleHealthCheckTimeoutOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.HealthCheckTimeout != 0 {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.HealthCheckTimeout = overrides.HealthCheckTimeout
		} else {
			app := manifest.GetFirstApp()
			app.HealthCheckTimeout = overrides.HealthCheckTimeout
		}
	}

	return manifest, nil
}
