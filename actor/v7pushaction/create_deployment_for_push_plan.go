package v7pushaction

func (actor Actor) CreateDeploymentForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Plan: pushPlan, Event: StartingDeployment}

	deploymentGUID, warnings, err := actor.V7Actor.CreateDeployment(pushPlan.Application.GUID, pushPlan.DropletGUID)

	if err != nil {
		return pushPlan, Warnings(warnings), err
	}

	eventStream <- &PushEvent{Plan: pushPlan, Event: WaitingForDeployment}

	handleInstanceDetails := func(instanceDetails string) {
		eventStream <- &PushEvent{
			Plan:     pushPlan,
			Warnings: Warnings{instanceDetails},
			Event:    InstanceDetails,
		}
	}

	pollWarnings, err := actor.V7Actor.PollStartForRolling(pushPlan.Application, deploymentGUID, pushPlan.NoWait, handleInstanceDetails)
	warnings = append(warnings, pollWarnings...)

	return pushPlan, Warnings(warnings), err
}
