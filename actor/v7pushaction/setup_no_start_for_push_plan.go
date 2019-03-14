package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupNoStartForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.NoStart = overrides.NoStart

	return pushPlan, nil
}
