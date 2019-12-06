package v7pushaction

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

// Overrides the path if path is given. Changes empty paths to pwd. Validates paths
func HandleAppPathOverride(manifest manifestparser.Manifest, overrides FlagOverrides) (manifestparser.Manifest, error) {
	if overrides.ProvidedAppPath != "" {
		if manifest.ContainsMultipleApps() {
			return manifest, translatableerror.CommandLineArgsWithMultipleAppsError{}
		}

		app := manifest.GetFirstApp()

		if app.Docker != nil {
			return manifest, translatableerror.ArgumentManifestMismatchError{
				Arg:              "--path, -p",
				ManifestProperty: "docker",
			}
		}

		app.Path = overrides.ProvidedAppPath
	}

	for i := range manifest.Applications {
		if manifest.Applications[i].Path == "" {
			continue
		}

		var finalPath = manifest.Applications[i].Path
		if !filepath.IsAbs(finalPath) && overrides.ProvidedAppPath == "" {
			finalPath = filepath.Join(filepath.Dir(manifest.PathToManifest), finalPath)
		}

		finalPathNoLinks, err := filepath.EvalSymlinks(finalPath)

		if err != nil {
			if os.IsNotExist(err) {
				return manifest, manifestparser.InvalidManifestApplicationPathError{
					Path: finalPath,
				}
			}

			return manifest, err
		}

		manifest.Applications[i].Path = finalPathNoLinks
	}

	return manifest, nil
}
