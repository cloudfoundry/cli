package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupSkipRouteCreationForPushPlan(pushPlan PushPlan, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.SkipRouteCreation = pushPlan.Overrides.SkipRouteCreation

	return pushPlan, nil
}
