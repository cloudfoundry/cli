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
