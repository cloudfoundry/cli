package pushaction

import "code.cloudfoundry.org/cli/util/manifest"

func (*Actor) ReadManifest(pathToManifest string, pathsToVarsFiles []string) ([]manifest.Application, error) {
	// Cover method to make testing easier
	return manifest.ReadAndInterpolateManifest(pathToManifest, pathsToVarsFiles)
}
