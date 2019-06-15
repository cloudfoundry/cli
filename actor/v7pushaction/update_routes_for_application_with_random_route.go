package v7pushaction

func (actor Actor) UpdateRoutesForApplicationWithRandomRoute(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Event: CreatingAndMappingRoutes}
	warnings, err := actor.CreateAndMapRoute(pushPlan.OrgGUID, pushPlan.SpaceGUID, pushPlan.Application, RandomRoute)
	if err != nil {
		return pushPlan, warnings, err
	}
	eventStream <- &PushEvent{Plan: pushPlan, Event: CreatedRoutes}
	return pushPlan, warnings, err
}
