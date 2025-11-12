package v7pushaction

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/cf/errors"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/types"
)

func SetupDeploymentInformationForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.Strategy = overrides.Strategy

	if overrides.Strategy != constant.DeploymentStrategyDefault && overrides.MaxInFlight != nil {
		pushPlan.MaxInFlight = *overrides.MaxInFlight
	}

	if overrides.Strategy == constant.DeploymentStrategyCanary && overrides.InstanceSteps != nil {
		pushPlan.InstanceSteps = overrides.InstanceSteps
	}

	if overrides.Strategy != constant.DeploymentStrategyDefault && overrides.Instances.IsSet {
		pushPlan.Instances = overrides.Instances
	}

	if overrides.Strategy != constant.DeploymentStrategyDefault && overrides.Memory != "" {
		size, err := flag.ConvertToMb(overrides.Memory)
		if err != nil {
			return PushPlan{}, errors.New(err.Error())
		}
		pushPlan.MemoryInMB.Value = size
		pushPlan.MemoryInMB.IsSet = true
	}
	if overrides.Strategy != constant.DeploymentStrategyDefault && overrides.Disk != "" {
		size, err := flag.ConvertToMb(overrides.Disk)
		if err != nil {
			return PushPlan{}, errors.New(err.Error())
		}
		pushPlan.DiskInMB.Value = size
		pushPlan.DiskInMB.IsSet = true
	}

	if overrides.Strategy != constant.DeploymentStrategyDefault && overrides.LogRateLimit != "" {
		logRateLimit := flag.BytesWithUnlimited{}
		if err := logRateLimit.IsValidValue(overrides.LogRateLimit); err != nil {
			return PushPlan{}, errors.New(err.Error())
		}
		pushPlan.LogRateLimitInBPS = types.NullInt(logRateLimit)
	}
	return pushPlan, nil
}
