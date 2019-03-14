package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupScaleWebProcessForPushPlan(pushPlan PushPlan, manifestApp manifestparser.Application) (PushPlan, error) {
	if pushPlan.Overrides.Memory.IsSet || pushPlan.Overrides.Disk.IsSet || pushPlan.Overrides.Instances.IsSet {
		pushPlan.ScaleWebProcessNeedsUpdate = true

		pushPlan.ScaleWebProcess = v7action.Process{
			Type:       constant.ProcessTypeWeb,
			DiskInMB:   pushPlan.Overrides.Disk,
			Instances:  pushPlan.Overrides.Instances,
			MemoryInMB: pushPlan.Overrides.Memory,
		}
	}
	return pushPlan, nil
}
