package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupSkipRouteCreationForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.SkipRouteCreation = manifestApp.NoRoute
	pushPlan.NoRouteFlag = overrides.SkipRouteCreation

	return pushPlan, nil
}
