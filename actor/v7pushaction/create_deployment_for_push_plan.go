package v7pushaction

func (actor Actor) CreateDeploymentForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- StartingDeployment

	deploymentGUID, warnings, err := actor.V7Actor.CreateDeployment(pushPlan.Application.GUID, pushPlan.DropletGUID)

	if err != nil {
		return pushPlan, Warnings(warnings), err
	}

	eventStream <- WaitingForDeployment

	pollWarnings, err := actor.V7Actor.PollStartForRolling(pushPlan.Application.GUID, deploymentGUID)
	warnings = append(warnings, pollWarnings...)

	return pushPlan, Warnings(warnings), err
}
