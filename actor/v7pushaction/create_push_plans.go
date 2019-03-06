package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"os"
)

// We assume that all flag and argument and manifest combinations have been validated by this point
func (actor Actor) CreatePushPlans(appNameArg string, parser manifestparser.Parser, overrides FlagOverrides) ([]PushPlan, error) {
	var pushPlans []PushPlan

	for _, application := range getEligibleApplications(parser, appNameArg) {
		var err error

		application.Name = appNameArg

		applicationNeedsUpdate := false

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

		bitsPath, err := getBitsPath(overrides)
		if err != nil {
			return nil, err
		}

		pushPlans = append(pushPlans, PushPlan{
			Application:            application,
			ApplicationNeedsUpdate: applicationNeedsUpdate,
			BitsPath:               bitsPath,
		})
	}

	return pushPlans, nil
}

func getEligibleApplications(parser manifestparser.Parser, appNameArg string) []v7action.Application {
	return []v7action.Application{{}}
}

func buildpacksPresent(overrides FlagOverrides) bool {
	return len(overrides.Buildpacks) > 0
}

func stacksPresent(overrides FlagOverrides) bool {
	return overrides.Stack != ""
}

func getBitsPath(overrides FlagOverrides) (string, error) {
	if overrides.ProvidedAppPath != "" {
		return overrides.ProvidedAppPath, nil
	}

	return os.Getwd()
}
