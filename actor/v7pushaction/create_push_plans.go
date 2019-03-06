package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"os"
)

// We assume that all flag and argument and manifest combinations have been validated by this point
func (actor Actor) CreatePushPlans(appNameArg string, spaceGUID string, orgGUID string, parser manifestparser.ManifestParser, overrides FlagOverrides) ([]PushPlan, error) {
	var pushPlans []PushPlan

	eligibleApps, err := getEligibleApplications(parser, appNameArg)
	if err != nil {
		return nil, err
	}
	for _, manifestApplication := range eligibleApps {
		var err error
		application := v7action.Application{Name: manifestApplication.Name}

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

		bitsPath, err := getBitsPath(overrides, manifestApplication.Path)
		if err != nil {
			return nil, err
		}

		pushPlans = append(pushPlans, PushPlan{
			OrgGUID:                orgGUID,
			SpaceGUID:              spaceGUID,
			Application:            application,
			ApplicationNeedsUpdate: applicationNeedsUpdate,
			BitsPath:               bitsPath,
		})
	}

	return pushPlans, nil
}

func getEligibleApplications(parser manifestparser.ManifestParser, appNameArg string) ([]manifestparser.Application, error) {
	if parser.FullRawManifest() != nil {
		return parser.Apps(appNameArg)
	}
	manifestApp := manifestparser.Application{}
	manifestApp.Name = appNameArg
	return []manifestparser.Application{manifestApp}, nil
}

func buildpacksPresent(overrides FlagOverrides) bool {
	return len(overrides.Buildpacks) > 0
}

func stacksPresent(overrides FlagOverrides) bool {
	return overrides.Stack != ""
}

func getBitsPath(overrides FlagOverrides, manifestPath string) (string, error) {
	if overrides.ProvidedAppPath != "" {
		return overrides.ProvidedAppPath, nil
	} else if manifestPath != "" {
		return manifestPath, nil
	}

	return os.Getwd()
}
