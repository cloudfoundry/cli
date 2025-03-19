package v7pushaction

import (
	"code.cloudfoundry.org/cli/v7/command/translatableerror"
	"code.cloudfoundry.org/cli/v7/util/manifestparser"
)

func HandleStrategyOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Strategy != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}
	}

	return manifest, nil
}
