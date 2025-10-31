package v7pushaction

import (
    "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/util/manifestparser"
)

func HandleInstancesOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Strategy != "" {
		return manifest, nil
	}

	if overrides.Instances.IsSet {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.Instances = &overrides.Instances.Value
		} else {
			app := manifest.GetFirstApp()
			app.Instances = &overrides.Instances.Value
		}
	}

	return manifest, nil
}

func HandleInstancesOverrideForDeployment(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Strategy != constant.DeploymentStrategyRolling && overrides.Strategy != constant.DeploymentStrategyCanary {
		return manifest, nil
	}

	if overrides.Instances.IsSet {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.Instances = &overrides.Instances.Value
		} else {
			app := manifest.GetFirstApp()
			app.Instances = &overrides.Instances.Value
		}
	}

	return manifest, nil
}
