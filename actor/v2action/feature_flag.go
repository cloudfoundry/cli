package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

type FeatureFlagState string

const (
	FeatureFlagEnabled  FeatureFlagState = "enabled"
	FeatureFlagDisabled FeatureFlagState = "disabled"
)

type FeatureFlag ccv2.FeatureFlag

func (f FeatureFlag) State() FeatureFlagState {
	if f.Enabled {
		return FeatureFlagEnabled
	}
	return FeatureFlagDisabled
}

func (actor Actor) GetFeatureFlags() ([]FeatureFlag, Warnings, error) {
	featureFlags, warnings, err := actor.CloudControllerClient.GetConfigFeatureFlags()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var convertedFeatureFlags []FeatureFlag
	for _, flag := range featureFlags {
		convertedFeatureFlags = append(convertedFeatureFlags, FeatureFlag(flag))
	}

	return convertedFeatureFlags, Warnings(warnings), nil
}
