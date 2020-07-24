package pushaction

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/cloudfoundry/bosh-cli/director/template"
)

func (actor *Actor) ReadManifest(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]manifest.Application, Warnings, error) {
	// Cover method to make testing easier
	apps, err := manifest.ReadAndInterpolateManifest(pathToManifest, pathsToVarsFiles, vars)
	warnings := actor.checkForBuildpack(apps)
	return apps, warnings, err
}

func (*Actor) checkForBuildpack(manifestApp []manifest.Application) Warnings {
	for _, app := range manifestApp {
		if app.Buildpack.IsSet {
			return Warnings{"Deprecation warning: Use of 'buildpack' attribute in manifest is deprecated in favor of 'buildpacks'. Please see https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#deprecated for alternatives and other app manifest deprecations. This feature will be removed in the future."}
		}
	}

	return nil
}
