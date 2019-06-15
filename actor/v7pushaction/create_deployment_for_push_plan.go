package v7pushaction

func (actor Actor) CreateDeploymentForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Plan: pushPlan, Event: StartingDeployment}

	deploymentGUID, warnings, err := actor.V7Actor.CreateDeployment(pushPlan.Application.GUID, pushPlan.DropletGUID)

	if err != nil {
		return pushPlan, Warnings(warnings), err
	}

	eventStream <- &PushEvent{Plan: pushPlan, Event: WaitingForDeployment}

	pollWarnings, err := actor.V7Actor.PollStartForRolling(pushPlan.Application.GUID, deploymentGUID, pushPlan.NoWait)
	warnings = append(warnings, pollWarnings...)

	return pushPlan, Warnings(warnings), err
}
