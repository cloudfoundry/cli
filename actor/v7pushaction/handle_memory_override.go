package v7pushaction

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func HandleMemoryOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Strategy != "" {
		return manifest, nil
	}

	if overrides.Memory != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.Memory = overrides.Memory
		} else {
			app := manifest.GetFirstApp()
			app.Memory = overrides.Memory
		}
	}

	return manifest, nil
}

func HandleMemoryOverrideForDeployment(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Strategy != constant.DeploymentStrategyRolling && overrides.Strategy != constant.DeploymentStrategyCanary {
		return manifest, nil
	}

	if overrides.Memory != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.Memory = overrides.Memory
		} else {
			app := manifest.GetFirstApp()
			app.Memory = overrides.Memory
		}
	}

	return manifest, nil
}
