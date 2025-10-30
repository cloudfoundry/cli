package v7pushaction

import (
    "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/util/manifestparser"
)

func HandleLogRateLimitOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Strategy != "" {
		return manifest, nil
	}

	if overrides.LogRateLimit != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.LogRateLimit = overrides.LogRateLimit
		} else {
			app := manifest.GetFirstApp()
			app.LogRateLimit = overrides.LogRateLimit
		}
	}

	return manifest, nil
}

func HandleLogRateLimitOverrideForDeployment(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Strategy != constant.DeploymentStrategyRolling && overrides.Strategy != constant.DeploymentStrategyCanary {
		return manifest, nil
	}

	if overrides.LogRateLimit != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.LogRateLimit = overrides.LogRateLimit
		} else {
			app := manifest.GetFirstApp()
			app.LogRateLimit = overrides.LogRateLimit
		}
	}

	return manifest, nil
}
