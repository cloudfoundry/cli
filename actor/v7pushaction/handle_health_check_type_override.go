package v7pushaction

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func HandleHealthCheckTypeOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.HealthCheckType != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.HealthCheckType = overrides.HealthCheckType
			if webProcess.HealthCheckType != constant.HTTP {
				webProcess.HealthCheckEndpoint = ""
			}
		} else {
			app := manifest.GetFirstApp()
			app.HealthCheckType = overrides.HealthCheckType
			if app.HealthCheckType != constant.HTTP {
				app.HealthCheckEndpoint = ""
			}
		}
	}

	return manifest, nil
}
