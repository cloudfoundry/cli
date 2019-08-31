package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func HandleStartCommandOverride(manifest pushmanifestparser.Manifest, overrides FlagOverrides) (pushmanifestparser.Manifest, error) {
	if overrides.StartCommand.IsSet {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.SetStartCommand(overrides.StartCommand.Value)
		} else {
			app := manifest.GetFirstApp()
			app.SetStartCommand(overrides.StartCommand.Value)
		}
	}

	return manifest, nil
}
