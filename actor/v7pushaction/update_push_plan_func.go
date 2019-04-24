package v7pushaction

import "code.cloudfoundry.org/cli/util/manifestparser"

// UpdatePushPlanFunc is a function that is used by CreatePushPlans to setup
// push plans for the push command.
type UpdatePushPlanFunc func(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error)
