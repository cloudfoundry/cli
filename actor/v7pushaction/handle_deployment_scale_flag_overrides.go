package v7pushaction

import (
	"code.cloudfoundry.org/cli/v9/util/manifestparser"
)

func (actor Actor) HandleDeploymentScaleFlagOverrides(
	baseManifest manifestparser.Manifest,
	flagOverrides FlagOverrides,
) (manifestparser.Manifest, error) {
	newManifest := baseManifest

	for _, transformPlan := range actor.TransformManifestSequenceForDeployment {
		var err error
		newManifest, err = transformPlan(newManifest, flagOverrides)
		if err != nil {
			return manifestparser.Manifest{}, err
		}
	}

	return newManifest, nil
}
