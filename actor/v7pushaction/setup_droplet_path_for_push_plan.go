package v7pushaction

func SetupDropletPathForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.DropletPath = overrides.DropletPath

	return pushPlan, nil
}
