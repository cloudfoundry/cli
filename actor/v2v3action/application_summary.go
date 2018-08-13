package v2v3action

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type ApplicationSummary struct {
	v3action.ApplicationSummary
	Routes                       v2action.Routes
	ApplicationInstanceWithStats []v2action.ApplicationInstanceWithStats
}

func (summary ApplicationSummary) GetIsolationSegmentName() (string, bool) {
	if len(summary.ApplicationInstanceWithStats) > 0 && len(summary.ApplicationInstanceWithStats[0].IsolationSegment) > 0 {
		return summary.ApplicationInstanceWithStats[0].IsolationSegment, true
	}
	return "", false
}

func (actor Actor) GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool) (ApplicationSummary, Warnings, error) {

	var summary ApplicationSummary
	var allWarnings Warnings

	v3Summary, warnings, err := actor.V3Actor.GetApplicationSummaryByNameAndSpace(appName, spaceGUID, withObfuscatedValues)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}
	v3Summary.ProcessSummaries.Sort()
	summary.ApplicationSummary = v3Summary

	routes, routeWarnings, err := actor.V2Actor.GetApplicationRoutes(summary.GUID)
	allWarnings = append(allWarnings, routeWarnings...)

	// Depending on the version of CC API - a V3 app can exist but the v2 API
	// cannot find it. So the code ignores the error and continues on.
	if _, ok := err.(ccerror.ResourceNotFoundError); err != nil && !ok {
		return ApplicationSummary{}, allWarnings, err
	}
	summary.Routes = routes

	if summary.State == constant.ApplicationStarted {
		appStats, warnings, err := actor.V2Actor.GetApplicationInstancesWithStatsByApplication(summary.GUID)
		allWarnings = append(allWarnings, warnings...)
		if _, ok := err.(ccerror.ResourceNotFoundError); err != nil && !ok {
			return ApplicationSummary{}, allWarnings, err
		}
		summary.ApplicationInstanceWithStats = appStats
	}

	return summary, allWarnings, nil
}
