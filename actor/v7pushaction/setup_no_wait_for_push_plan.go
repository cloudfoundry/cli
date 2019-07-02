package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupNoWaitForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.NoWait = overrides.NoWait

	return pushPlan, nil
}
