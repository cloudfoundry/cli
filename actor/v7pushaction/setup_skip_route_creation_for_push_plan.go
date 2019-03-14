package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupSkipRouteCreationForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.SkipRouteCreation = overrides.SkipRouteCreation

	return pushPlan, nil
}
