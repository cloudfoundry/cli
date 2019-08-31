package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/pushmanifestparser"
)

func (actor Actor) HandleFlagOverrides(
	baseManifest pushmanifestparser.Manifest,
	flagOverrides FlagOverrides,
) (pushmanifestparser.Manifest, error) {
	newManifest := baseManifest

	for _, transformPlan := range actor.TransformManifestSequence {
		var err error
		newManifest, err = transformPlan(newManifest, flagOverrides)
		if err != nil {
			return pushmanifestparser.Manifest{}, err
		}
	}

	return newManifest, nil
}
