package pushaction

import (
	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/actor/sharedaction"
	"code.cloudfoundry.org/cli/v7/actor/v3action"
)

type PushPlan struct {
	Application        v3action.Application
	SpaceGUID          string
	BitsPath           string
	AllResources       []sharedaction.Resource
	MatchedResources   []sharedaction.Resource
	UnmatchedResources []sharedaction.Resource
	Archive            bool
}

func (actor Actor) Conceptualize(settings CommandLineSettings, spaceGUID string) ([]PushPlan, Warnings, error) {
	var (
		application v3action.Application
		warnings    v3action.Warnings
		err         error
	)

	application, warnings, err = actor.V3Actor.GetApplicationByNameAndSpace(settings.Name, spaceGUID)
	if _, ok := err.(actionerror.ApplicationNotFoundError); ok {
		application = v3action.Application{
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

	desiredState := []PushPlan{
		{
			Application:  application,
			SpaceGUID:    spaceGUID,
			BitsPath:     bitsPath,
			AllResources: resources,
		},
	}
	return desiredState, Warnings(warnings), err
}
