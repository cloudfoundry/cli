package v7pushaction

// UpdatePushPlanFunc is a function that is used by CreatePushPlans to setup
// push plans for the push command.
type UpdatePushPlanFunc func(pushPlan PushPlan, overrides FlagOverrides) (PushPlan, error)
