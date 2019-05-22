package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"errors"
	"os"
)

func (actor Actor) SetupAllResourcesForPushPlan(pushPlan PushPlan, overrides FlagOverrides, manifestApp manifestparser.Application) (PushPlan, error) {
	if pushPlan.DropletPath != "" {
		return pushPlan, nil
	}

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

	var v3Resources []sharedaction.V3Resource
	for _, resource := range resources {
		v3Resources = append(v3Resources, resource.ToV3Resource())
	}

	pushPlan.Archive = archive
	pushPlan.AllResources = v3Resources

	return pushPlan, nil
}
