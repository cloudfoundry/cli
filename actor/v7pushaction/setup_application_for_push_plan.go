package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

func SetupApplicationForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	application := v7action.Application{Name: manifestApp.Name}

	var applicationNeedsUpdate bool

	if buildpacksPresent(overrides) {
		application.LifecycleType = constant.AppLifecycleTypeBuildpack
		application.LifecycleBuildpacks = overrides.Buildpacks
		applicationNeedsUpdate = true
	}

	if stacksPresent(overrides) {
		application.StackName = overrides.Stack
		application.LifecycleType = constant.AppLifecycleTypeBuildpack
		applicationNeedsUpdate = true
	}

	if overrides.DockerImage != "" {
		application.LifecycleType = constant.AppLifecycleTypeDocker
	}

	pushPlan.Application = application
	pushPlan.ApplicationNeedsUpdate = applicationNeedsUpdate

	return pushPlan, nil
}

func buildpacksPresent(overrides FlagOverrides) bool {
	return len(overrides.Buildpacks) > 0
}

func stacksPresent(overrides FlagOverrides) bool {
	return overrides.Stack != ""
}
