package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func (actor Actor) HandleFlagOverrides(
	baseManifest manifestparser.Manifest,
	flagOverrides FlagOverrides,
) (manifestparser.Manifest, error) {
	newManifest := baseManifest

	for _, transformPlan := range actor.TransformManifestSequence {
		var err error
		newManifest, err = transformPlan(newManifest, flagOverrides)
		if err != nil {
			return manifestparser.Manifest{}, err
		}
	}

	return newManifest, nil
}
