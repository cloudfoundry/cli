package v7action

import (
	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v7/resources"
)

// GetFeatureFlagByName returns a featureFlag with the provided name.
func (actor Actor) GetFeatureFlagByName(featureFlagName string) (resources.FeatureFlag, Warnings, error) {
	var (
		featureFlags resources.FeatureFlag
		warnings     ccv3.Warnings
		err          error
	)
	featureFlags, warnings, err = actor.CloudControllerClient.GetFeatureFlag(featureFlagName)

	if err != nil {
		if _, ok := err.(ccerror.FeatureFlagNotFoundError); ok {
			return resources.FeatureFlag{}, Warnings(warnings), actionerror.FeatureFlagNotFoundError{FeatureFlagName: featureFlagName}
		}
		return resources.FeatureFlag{}, Warnings(warnings), err
	}

	return featureFlags, Warnings(warnings), err
}

func (actor Actor) GetFeatureFlags() ([]resources.FeatureFlag, Warnings, error) {

	var (
		featureFlags []resources.FeatureFlag
	)
	featureFlags, warnings, err := actor.CloudControllerClient.GetFeatureFlags()

	if err != nil {
		return nil, Warnings(warnings), err
	}

	return featureFlags, Warnings(warnings), nil
}

func (actor Actor) EnableFeatureFlag(flagName string) (Warnings, error) {
	return actor.updateFeatureFlag(resources.FeatureFlag{Name: flagName, Enabled: true})
}

func (actor Actor) DisableFeatureFlag(flagName string) (Warnings, error) {
	return actor.updateFeatureFlag(resources.FeatureFlag{Name: flagName, Enabled: false})
}

func (actor Actor) updateFeatureFlag(flag resources.FeatureFlag) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.UpdateFeatureFlag(flag)

	if _, ok := err.(ccerror.FeatureFlagNotFoundError); ok {
		err = actionerror.FeatureFlagNotFoundError{FeatureFlagName: flag.Name}
	}
	return Warnings(warnings), err
}
