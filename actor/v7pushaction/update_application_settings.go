package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"

	log "github.com/sirupsen/logrus"
)

// UpdateApplicationSettings syncs the Application state and GUID with the API.
func (actor Actor) UpdateApplicationSettings(pushPlans []PushPlan) ([]PushPlan, Warnings, error) {
	appNames := actor.getAppNames(pushPlans)
	applications, getWarnings, err := actor.V7Actor.GetApplicationsByNamesAndSpace(appNames, pushPlans[0].SpaceGUID)
	warnings := Warnings(getWarnings)
	if err != nil {
		log.Errorln("Looking up applications:", err)
		return nil, warnings, err
	}

	nameToApp := actor.generateAppNameToApplicationMapping(applications)

	var updatedPushPlans []PushPlan
	for _, pushPlan := range pushPlans {
		app := nameToApp[pushPlan.Application.Name]
		pushPlan.Application.GUID = app.GUID
		pushPlan.Application.State = app.State

		routes, getWarnings, err := actor.V7Actor.GetApplicationRoutes(app.GUID)
		warnings = append(warnings, Warnings(getWarnings)...)
		if err != nil {
			log.Errorln("Retrieving routes:", err)
			return nil, warnings, err
		}

		pushPlan.ApplicationRoutes = routes

		updatedPushPlans = append(updatedPushPlans, pushPlan)
	}

	return updatedPushPlans, warnings, err
}

func (Actor) getAppNames(pushPlans []PushPlan) []string {
	var appNames []string
	for _, pushPlan := range pushPlans {
		appNames = append(appNames, pushPlan.Application.Name)
	}
	return appNames
}

func (Actor) generateAppNameToApplicationMapping(applications []v7action.Application) map[string]v7action.Application {
	nameToApp := map[string]v7action.Application{}
	for _, currApp := range applications {
		nameToApp[currApp.Name] = currApp
	}
	return nameToApp
}
