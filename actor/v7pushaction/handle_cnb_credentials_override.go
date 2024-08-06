package v7pushaction

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func HandleCNBCredentialsOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.CNBCredentials != nil {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()

		if app.RemainingManifestFields == nil {
			app.RemainingManifestFields = map[string]interface{}{}
		}

		app.RemainingManifestFields["cnb-credentials"] = overrides.CNBCredentials
	}

	return manifest, nil
}
