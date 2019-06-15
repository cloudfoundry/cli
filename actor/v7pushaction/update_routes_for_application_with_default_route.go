package v7pushaction

func (actor Actor) UpdateRoutesForApplicationWithDefaultRoute(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Event: CreatingAndMappingRoutes}
	warnings, err := actor.CreateAndMapRoute(pushPlan.OrgGUID, pushPlan.SpaceGUID, pushPlan.Application, DefaultRoute)
	if err != nil {
		return pushPlan, warnings, err
	}
	eventStream <- &PushEvent{Plan: pushPlan, Event: CreatedRoutes}
	return pushPlan, warnings, err
}
