package v7pushaction

import (
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/util/manifestparser"
)

func HandleDiskOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.Disk != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		webProcess := manifest.GetFirstAppWebProcess()
		if webProcess != nil {
			webProcess.DiskQuota = overrides.Disk
		} else {
			app := manifest.GetFirstApp()
			app.DiskQuota = overrides.Disk
		}
	}

	return manifest, nil
}
