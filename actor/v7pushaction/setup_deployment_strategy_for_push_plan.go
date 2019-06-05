package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupDeploymentStrategyForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.Strategy = overrides.Strategy

	return pushPlan, nil
}
