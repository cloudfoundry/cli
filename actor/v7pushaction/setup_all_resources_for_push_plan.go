package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"errors"
	"os"
)

func (actor Actor) SetupAllResourcesForPushPlan(pushPlan PushPlan, manifestApp manifestparser.Application) (PushPlan, error) {
	if pushPlan.Application.LifecycleType == constant.AppLifecycleTypeDocker {
		return pushPlan, nil
	}

	path := pushPlan.BitsPath
	if path == "" {
		return PushPlan{}, errors.New("developer error: Bits Path needs to be set prior to generating app resources")
	}

	info, err := os.Stat(path)
	if err != nil {
		return PushPlan{}, err
	}

	var archive bool
	var resources []sharedaction.Resource
	if info.IsDir() {
		resources, err = actor.SharedActor.GatherDirectoryResources(path)
	} else {
		archive = true
		resources, err = actor.SharedActor.GatherArchiveResources(path)
	}
	if err != nil {
		return PushPlan{}, err
	}

	pushPlan.Archive = archive
	pushPlan.AllResources = resources

	return pushPlan, nil
}
