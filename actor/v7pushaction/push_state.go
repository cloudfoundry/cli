package v7pushaction

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/types"
)

type PushState struct {
	Application v7action.Application
	SpaceGUID   string
	OrgGUID     string
	Overrides   FlagOverrides

	Archive            bool
	BitsPath           string
	AllResources       []sharedaction.Resource
	MatchedResources   []sharedaction.Resource
	UnmatchedResources []sharedaction.Resource
}

type FlagOverrides struct {
	ProvidedAppPath string
	Memory          types.NullUint64
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

func (actor Actor) Conceptualize(appName string, spaceGUID string, orgGUID string, currentDir string, flagOverrides FlagOverrides) ([]PushState, Warnings, error) {
	var (
		application v7action.Application
		warnings    v7action.Warnings
		err         error
	)

	application, warnings, err = actor.V7Actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if _, ok := err.(actionerror.ApplicationNotFoundError); ok {
		application = v7action.Application{
			Name: appName,
		}
	} else if err != nil {
		return nil, Warnings(warnings), err
	}

	bitsPath := currentDir
	if flagOverrides.ProvidedAppPath != "" {
		bitsPath = flagOverrides.ProvidedAppPath
	}

	resources, err := actor.SharedActor.GatherDirectoryResources(bitsPath)

	desiredState := []PushState{
		{
			Application: application,
			SpaceGUID:   spaceGUID,
			OrgGUID:     orgGUID,
			Overrides:   flagOverrides,

			BitsPath:     bitsPath,
			AllResources: resources,
		},
	}
	return desiredState, Warnings(warnings), err
}
