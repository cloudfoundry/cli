package v7pushaction

func (actor Actor) StagePackageForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Plan: pushPlan, Event: StartingStaging}

	var allWarnings Warnings
	build, warnings, err := actor.V7Actor.StageApplicationPackage(pushPlan.PackageGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return pushPlan, allWarnings, err
	}

	eventStream <- &PushEvent{Plan: pushPlan, Event: PollingBuild}

	droplet, warnings, err := actor.V7Actor.PollBuild(build.GUID, pushPlan.Application.Name)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return pushPlan, allWarnings, err
	}
	pushPlan.DropletGUID = droplet.GUID

	eventStream <- &PushEvent{Plan: pushPlan, Event: StagingComplete}

	return pushPlan, allWarnings, nil
}
