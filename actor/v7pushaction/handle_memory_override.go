package v7pushaction

import (
	//"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func HandleMemoryOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
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
