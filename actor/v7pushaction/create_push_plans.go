package v7pushaction

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
)

// CreatePushPlans returns a set of PushPlan objects based off the inputs
// provided. It's assumed that all flag and argument and manifest combinations
// have been validated prior to calling this function.
func (actor Actor) CreatePushPlans(
	appNameArg string,
	spaceGUID string,
	orgGUID string,
	parser ManifestParser,
	overrides FlagOverrides,
) ([]PushPlan, error) {
	var pushPlans []PushPlan

	for _, manifestApplication := range getEligibleApplications(parser, appNameArg) {
		plan := PushPlan{
			OrgGUID:   orgGUID,
			SpaceGUID: spaceGUID,
		}

		// List of PreparePushPlanSequence is defined in NewActor
		for _, updatePlan := range actor.PreparePushPlanSequence {
			var err error
			plan, err = updatePlan(plan, overrides, manifestApplication)
			if err != nil {
				return nil, err
			}
		}

		pushPlans = append(pushPlans, plan)
	}

	return pushPlans, nil
}

func getEligibleApplications(parser ManifestParser, appName string) []manifestparser.Application {
	if parser.ContainsManifest() {
		return parser.Apps()
	}
	manifestApp := manifestparser.Application{}
	manifestApp.Name = appName
	return []manifestparser.Application{manifestApp}
}
