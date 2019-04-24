package v7pushaction



func (actor Actor) UpdateRoutesForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	if !(pushPlan.SkipRouteCreation || pushPlan.NoRouteFlag) {
		eventStream <- CreatingAndMappingRoutes
		warnings, err := actor.CreateAndMapDefaultApplicationRoute(pushPlan.OrgGUID, pushPlan.SpaceGUID, pushPlan.Application)

		if err != nil {
			return pushPlan, warnings, err
		}
		eventStream <- CreatedRoutes
		return pushPlan, Warnings(warnings), err
	}
	return pushPlan, nil, nil
}
