package v7pushaction

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

func SetupDeploymentInformationForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.Strategy = overrides.Strategy

	if overrides.Strategy != constant.DeploymentStrategyDefault && overrides.MaxInFlight != nil {
		pushPlan.MaxInFlight = *overrides.MaxInFlight
	}

	if overrides.Strategy == constant.DeploymentStrategyCanary && overrides.InstanceSteps != nil {
		pushPlan.InstanceSteps = overrides.InstanceSteps
	}

	return pushPlan, nil
}
