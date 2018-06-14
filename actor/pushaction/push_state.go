package pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v3action"
)

type PushState struct {
	Application v3action.Application
	SpaceGUID   string
	// Path        string //TODO: more descriptive name - feels ambiguous - is this a manifest path or app dir path?

	// AllResources       []v2action.Resource
	// MatchedResources   []v2action.Resource
	// UnmatchedResources []v2action.Resource
	// Archive bool
}

func (actor Actor) GeneratePushState(settings CommandLineSettings, spaceGUID string) ([]PushState, Warnings, error) {
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

	desiredState := []PushState{
		{
			Application: application,
			SpaceGUID:   spaceGUID,
			// Path:        settings.ProvidedAppPath,
		},
	}
	return desiredState, Warnings(warnings), nil
}
