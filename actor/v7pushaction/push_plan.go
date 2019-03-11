package v7pushaction

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	log "github.com/sirupsen/logrus"
)

type PushPlan struct {
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
	HealthCheckTimeout  int64
	HealthCheckType     constant.HealthCheckType
	Instances           types.NullInt
	Memory              types.NullUint64
	NoStart             bool
	ProvidedAppPath     string
	SkipRouteCreation   bool
	StartCommand        types.FilteredString
}

func (state PushPlan) String() string {
	return fmt.Sprintf(
		"Application: %#v - Space GUID: %s, Org GUID: %s, Archive: %t, Bits Path: %s",
		state.Application,
		state.SpaceGUID,
		state.OrgGUID,
		state.Archive,
		state.BitsPath,
	)
}

// Update Application for push plan
func (actor Actor) Conceptualize(pushPlans []PushPlan) ([]PushPlan, Warnings, error) {
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
