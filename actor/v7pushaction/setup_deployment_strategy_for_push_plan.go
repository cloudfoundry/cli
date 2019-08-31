package v7pushaction

func SetupDeploymentStrategyForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.Strategy = overrides.Strategy

	return pushPlan, nil
}
