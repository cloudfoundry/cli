package pushaction

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/cloudfoundry/bosh-cli/director/template"
)

func (*Actor) ReadManifest(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]manifest.Application, error) {
	// Cover method to make testing easier
	return manifest.ReadAndInterpolateManifest(pathToManifest, pathsToVarsFiles, vars)
}
