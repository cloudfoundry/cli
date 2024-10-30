package v7pushaction

import (
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/util/manifestparser"
)

func HandleStackOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Stack != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}
		manifest.Applications[0].Stack = overrides.Stack
	}

	return manifest, nil
}
