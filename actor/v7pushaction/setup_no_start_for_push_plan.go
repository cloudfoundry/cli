package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupNoStartForPushPlan(pushPlan PushPlan, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.NoStart = pushPlan.Overrides.NoStart

	return pushPlan, nil
}
