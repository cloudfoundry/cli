package v7pushaction

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func HandleHealthCheckEndpointOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
	if overrides.HealthCheckEndpoint != "" {

		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		var healthCheckType constant.HealthCheckType

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.HealthCheckEndpoint = overrides.HealthCheckEndpoint
			healthCheckType = webProcess.HealthCheckType
		} else {
			app := manifest.GetFirstApp()
			app.HealthCheckEndpoint = overrides.HealthCheckEndpoint
			healthCheckType = app.HealthCheckType
		}

		if healthCheckType != "" && healthCheckType != constant.HTTP {
			return manifest, translatableerror.ArgumentManifestMismatchError{
				Arg:              "--endpoint",
				ManifestProperty: "health-check-type",
				ManifestValue:    string(healthCheckType),
			}
		}
	}

	return manifest, nil
}
