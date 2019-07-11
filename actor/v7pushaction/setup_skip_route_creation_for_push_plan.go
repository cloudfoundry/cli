package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupSkipRouteCreationForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.SkipRouteCreation = manifestApp.NoRoute || overrides.NoRoute || manifestApp.RandomRoute
	pushPlan.NoRouteFlag = overrides.NoRoute
	pushPlan.RandomRoute = overrides.RandomRoute

	return pushPlan, nil
}
