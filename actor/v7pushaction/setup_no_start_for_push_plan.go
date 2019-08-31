package v7pushaction

func SetupNoStartForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.NoStart = overrides.NoStart

	return pushPlan, nil
}
