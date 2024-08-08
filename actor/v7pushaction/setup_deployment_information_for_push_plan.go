package v7pushaction

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

func SetupDeploymentInformationForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.Strategy = overrides.Strategy

	if overrides.Strategy != constant.DeploymentStrategyDefault {
		pushPlan.MaxInFlight = overrides.MaxInFlight
	}

	return pushPlan, nil
}
