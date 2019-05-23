package v7pushaction

func (actor Actor) UpdateRoutesForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- CreatingAndMappingRoutes
	warnings, err := actor.CreateAndMapDefaultApplicationRoute(pushPlan.OrgGUID, pushPlan.SpaceGUID, pushPlan.Application)
	if err != nil {
		return pushPlan, warnings, err
	}
	eventStream <- CreatedRoutes
	return pushPlan, warnings, err
}
