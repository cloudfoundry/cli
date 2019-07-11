package v7pushaction

func (actor Actor) UpdateRoutesForApplicationWithRandomRoute(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- CreatingAndMappingRoutes
	warnings, err := actor.CreateAndMapRoute(pushPlan.OrgGUID, pushPlan.SpaceGUID, pushPlan.Application, RandomRoute)
	if err != nil {
		return pushPlan, warnings, err
	}
	eventStream <- CreatedRoutes
	return pushPlan, warnings, err
}
