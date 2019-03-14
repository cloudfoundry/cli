package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupScaleWebProcessForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	if overrides.Memory.IsSet || overrides.Disk.IsSet || overrides.Instances.IsSet {
		pushPlan.ScaleWebProcessNeedsUpdate = true

		pushPlan.ScaleWebProcess = v7action.Process{
			Type:       constant.ProcessTypeWeb,
			DiskInMB:   overrides.Disk,
			Instances:  overrides.Instances,
			MemoryInMB: overrides.Memory,
		}
	}
	return pushPlan, nil
}
