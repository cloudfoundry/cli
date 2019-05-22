package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupDropletPathForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	pushPlan.DropletPath = overrides.DropletPath

	return pushPlan, nil
}
