package v7pushaction

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	log "github.com/sirupsen/logrus"
)

type PushState struct {
	Application v7action.Application
	SpaceGUID   string
	OrgGUID     string
	Overrides   FlagOverrides
	Manifest    []byte

	Archive                bool
	ApplicationNeedsUpdate bool
	BitsPath               string
	AllResources           []sharedaction.Resource
	MatchedResources       []sharedaction.Resource
	UnmatchedResources     []sharedaction.Resource
}

type FlagOverrides struct {
	Buildpacks          []string
	Stack               string
	Disk                types.NullUint64
	DockerImage         string
	DockerPassword      string
	DockerUsername      string
	HealthCheckEndpoint string
	HealthCheckType     constant.HealthCheckType
	Instances           types.NullInt
	Memory              types.NullUint64
	NoStart             bool
	ProvidedAppPath     string
	SkipRouteCreation   bool
	StartCommand        types.FilteredString
}

func (state PushState) String() string {
	return fmt.Sprintf(
		"Application: %#v - Space GUID: %s, Org GUID: %s, Archive: %t, Bits Path: %s",
		state.Application,
		state.SpaceGUID,
		state.OrgGUID,
		state.Archive,
		state.BitsPath,
	)
}

func sliceContains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func (actor Actor) Conceptualize(
	appNames []string,
	spaceGUID string,
	orgGUID string,
	currentDir string,
	flagOverrides FlagOverrides,
) ([]PushState, Warnings, error) {
	var (
		applications []v7action.Application
		warnings     v7action.Warnings
		err          error
	)
	// for each appName get the app
	var getWarnings v7action.Warnings
	applications, getWarnings, err = actor.V7Actor.GetApplicationsByNamesAndSpace(appNames, spaceGUID)
	warnings = append(warnings, getWarnings...)
	if err != nil {
		log.Errorln("Looking up applications:", err)
		return nil, Warnings(warnings), err
	}

	pushStates := []PushState{}
	for _, application := range applications {
		applicationNeedsUpdate := false

		if len(flagOverrides.Buildpacks) != 0 {
			application.LifecycleType = constant.AppLifecycleTypeBuildpack
			application.LifecycleBuildpacks = flagOverrides.Buildpacks
			applicationNeedsUpdate = true
		}

		if flagOverrides.Stack != "" {
			application.LifecycleType = constant.AppLifecycleTypeBuildpack
			application.StackName = flagOverrides.Stack
			applicationNeedsUpdate = true
		}

		bitsPath := currentDir
		if flagOverrides.ProvidedAppPath != "" {
			bitsPath = flagOverrides.ProvidedAppPath
		}

		var info os.FileInfo
		info, err = os.Stat(bitsPath)
		if err != nil {
			return nil, Warnings(warnings), err
		}

		var archive bool
		var resources []sharedaction.Resource
		if info.IsDir() {
			resources, err = actor.SharedActor.GatherDirectoryResources(bitsPath)
		} else {
			archive = true
			resources, err = actor.SharedActor.GatherArchiveResources(bitsPath)
		}
		if err != nil {
			return nil, Warnings(warnings), err
		}

		pushStates = append(pushStates, PushState{
			Application: application,
			SpaceGUID:   spaceGUID,
			OrgGUID:     orgGUID,
			Overrides:   flagOverrides,

			Archive:                archive,
			BitsPath:               bitsPath,
			AllResources:           resources,
			ApplicationNeedsUpdate: applicationNeedsUpdate,
		})
	}

	return pushStates, Warnings(warnings), err
}
