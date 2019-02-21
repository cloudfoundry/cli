package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type FeatureFlag ccv3.FeatureFlag

// Returns a featureFlag with the provided name.
func (actor Actor) GetFeatureFlagByName(featureFlagName string) (FeatureFlag, Warnings, error) {
	var (
		ccv3FeatureFlag ccv3.FeatureFlag
		warnings        ccv3.Warnings
		err             error
	)
	ccv3FeatureFlag, warnings, err = actor.CloudControllerClient.GetFeatureFlag(featureFlagName)

	if err != nil {
		if _, ok := err.(ccerror.FeatureFlagNotFoundError); ok {
			return FeatureFlag{}, Warnings(warnings), actionerror.FeatureFlagNotFoundError{FeatureFlagName: featureFlagName}
		}
		return FeatureFlag{}, Warnings(warnings), err
	}

	return FeatureFlag(ccv3FeatureFlag), Warnings(warnings), err
}

func (actor Actor) GetFeatureFlags() ([]FeatureFlag, Warnings, error) {

	var (
		featureFlags []FeatureFlag
	)
	ccv3FeatureFlags, warnings, err := actor.CloudControllerClient.GetFeatureFlags()

	if err != nil {
		return nil, Warnings(warnings), err
	}

	for _, flag := range ccv3FeatureFlags {
		featureFlags = append(featureFlags, FeatureFlag(flag))
	}

	return featureFlags, Warnings(warnings), nil
}
