package pushaction

import "code.cloudfoundry.org/cli/util/manifest"

func (*Actor) ReadManifest(pathToManifest string) ([]manifest.Application, error) {
	// Cover method to make testing easier
	return manifest.ReadAndMergeManifests(pathToManifest)
}
