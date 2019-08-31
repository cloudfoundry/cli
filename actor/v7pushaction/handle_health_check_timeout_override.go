package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func HandleHealthCheckTimeoutOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
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
