package v7pushaction

func (actor Actor) CreateDockerPackageForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Plan: pushPlan, Event: SetDockerImage}

	pkg, warnings, err := actor.V7Actor.CreateDockerPackageByApplication(pushPlan.Application.GUID, pushPlan.DockerImageCredentials)
	if err != nil {
		return pushPlan, Warnings(warnings), err
	}

	eventStream <- &PushEvent{Plan: pushPlan, Event: SetDockerImageComplete}

	polledPackage, pollWarnings, err := actor.V7Actor.PollPackage(pkg)

	pushPlan.PackageGUID = polledPackage.GUID

	return pushPlan, Warnings(append(warnings, pollWarnings...)), err
}
