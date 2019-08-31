package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func HandleHealthCheckEndpointOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
	if overrides.HealthCheckEndpoint != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.HealthCheckEndpoint = overrides.HealthCheckEndpoint
		} else {
			app := manifest.GetFirstApp()
			app.HealthCheckEndpoint = overrides.HealthCheckEndpoint
		}
	}

	return manifest, nil
}
