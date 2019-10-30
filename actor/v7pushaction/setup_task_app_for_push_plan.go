package v7pushaction

func SetupTaskAppForPushPlan(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error) {
	pushPlan.TaskTypeApplication = overrides.Task

	return pushPlan, nil
}
