package pushaction

import "code.cloudfoundry.org/cli/actor/pushaction/manifest"

func (*Actor) ReadManifest(pathToManifest string) ([]manifest.Application, error) {
	// Cover method to make testing easier
	return manifest.ReadAndMergeManifests(pathToManifest)
}
