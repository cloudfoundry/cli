package v7pushaction

func SetupNoWaitForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.NoWait = overrides.NoWait

	return pushPlan, nil
}
