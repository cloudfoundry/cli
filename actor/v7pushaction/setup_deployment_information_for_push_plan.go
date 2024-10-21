package v7pushaction

import "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"

func SetupDeploymentInformationForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.Strategy = overrides.Strategy

	if overrides.Strategy != constant.DeploymentStrategyDefault && overrides.MaxInFlight != nil {
		pushPlan.MaxInFlight = *overrides.MaxInFlight
	}

	return pushPlan, nil
}
