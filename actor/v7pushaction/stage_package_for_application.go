package v7pushaction

func (actor Actor) StagePackageForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	if pushPlan.DropletPath != "" {
		return pushPlan, nil, nil
	}

	eventStream <- StartingStaging

	var allWarnings Warnings
	build, warnings, err := actor.V7Actor.StageApplicationPackage(pushPlan.PackageGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return pushPlan, allWarnings, err
	}

	eventStream <- PollingBuild

	droplet, warnings, err := actor.V7Actor.PollBuild(build.GUID, pushPlan.Application.Name)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return pushPlan, allWarnings, err
	}
	pushPlan.DropletGUID = droplet.GUID

	eventStream <- StagingComplete

	return pushPlan, allWarnings, nil
}
