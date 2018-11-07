package v7pushaction

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
)

type PushState struct {
	Application v7action.Application
	OrgGUID     string
	SpaceGUID   string

	Archive            bool
	BitsPath           string
	AllResources       []sharedaction.Resource
	MatchedResources   []sharedaction.Resource
	UnmatchedResources []sharedaction.Resource
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

func (actor Actor) Conceptualize(settings CommandLineSettings, spaceGUID string, orgGUID string) ([]PushState, Warnings, error) {
	var (
		application v7action.Application
		warnings    v7action.Warnings
		err         error
	)

	application, warnings, err = actor.V7Actor.GetApplicationByNameAndSpace(settings.Name, spaceGUID)
	if _, ok := err.(actionerror.ApplicationNotFoundError); ok {
		application = v7action.Application{
			Name: settings.Name,
		}
	} else if err != nil {
		return nil, Warnings(warnings), err
	}

	bitsPath := settings.CurrentDirectory
	if settings.ProvidedAppPath != "" {
		bitsPath = settings.ProvidedAppPath
	}

	resources, err := actor.SharedActor.GatherDirectoryResources(bitsPath)

	desiredState := []PushState{
		{
			Application:  application,
			SpaceGUID:    spaceGUID,
			OrgGUID:      orgGUID,
			BitsPath:     bitsPath,
			AllResources: resources,
		},
	}
	return desiredState, Warnings(warnings), err
}
