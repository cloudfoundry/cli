package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/manifestparser"
)

// CreatePushPlans returns a set of PushPlan objects based off the inputs
// provided. It's assumed that all flag and argument and manifest combinations
// have been validated prior to calling this function.
func (actor Actor) CreatePushPlans(
	spaceGUID string,
	orgGUID string,
	manifest manifestparser.Manifest,
	overrides FlagOverrides,
) ([]PushPlan, v7action.Warnings, error) {
	var pushPlans []PushPlan

	apps, warnings, err := actor.V7Actor.GetApplicationsByNamesAndSpace(manifest.AppNames(), spaceGUID)
	if err != nil {
		return nil, warnings, err
	}
	nameToApp := actor.generateAppNameToApplicationMapping(apps)

	for _, manifestApplication := range manifest.Applications {
		plan := PushPlan{
			OrgGUID:     orgGUID,
			SpaceGUID:   spaceGUID,
			Application: nameToApp[manifestApplication.Name],
			BitsPath:    manifestApplication.Path,
		}

		if manifestApplication.Lifecycle != "" {
			plan.Application.LifecycleType = manifestApplication.Lifecycle
		}

		if overrides.Lifecycle != "" {
			plan.Application.LifecycleType = overrides.Lifecycle
		}

		if overrides.CNBCredentials != nil {
			plan.Application.Credentials = overrides.CNBCredentials
		}

		if manifestApplication.Docker != nil {
			plan.DockerImageCredentials = v7action.DockerImageCredentials{
				Path:     manifestApplication.Docker.Image,
				Username: manifestApplication.Docker.Username,
				Password: overrides.DockerPassword,
			}
		}

		// List of PreparePushPlanSequence is defined in NewActor
		for _, updatePlan := range actor.PreparePushPlanSequence {
			var err error
			plan, err = updatePlan(plan, overrides)
			if err != nil {
				return nil, warnings, err
			}
		}

		pushPlans = append(pushPlans, plan)
	}

	return pushPlans, warnings, nil
}

func (actor Actor) generateAppNameToApplicationMapping(applications []resources.Application) map[string]resources.Application {
	nameToApp := make(map[string]resources.Application, len(applications))
	for _, app := range applications {
		nameToApp[app.Name] = app
	}
	return nameToApp
}
