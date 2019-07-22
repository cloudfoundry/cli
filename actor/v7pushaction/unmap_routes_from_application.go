package v7pushaction

func (actor Actor) UnmapRoutesFromApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	var allWarnings Warnings

	eventStream <- &PushEvent{Event: UnmappingRoutes}
	for _, route := range pushPlan.ApplicationRoutes {
		destination, warnings, err := actor.V7Actor.GetRouteDestinationByAppGUID(route.GUID, pushPlan.Application.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return pushPlan, allWarnings, err
		}

		warnings, err = actor.V7Actor.UnmapRoute(route.GUID, destination.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return pushPlan, allWarnings, err
		}
	}

	return pushPlan, allWarnings, nil
}
