package pushaction

import "code.cloudfoundry.org/cli/util/manifest"

func (*Actor) ReadManifest(pathToManifest string, pathToVarsFile string) ([]manifest.Application, error) {
	// Cover method to make testing easier
	return manifest.ReadAndInterpolateManifest(pathToManifest, pathToVarsFile)
}
