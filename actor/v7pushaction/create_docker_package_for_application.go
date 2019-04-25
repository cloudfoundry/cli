package v7pushaction

func (actor Actor) CreateDockerPackageForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	if pushPlan.DockerImageCredentialsNeedsUpdate {
		eventStream <- SetDockerImage

		pkg, warnings, err := actor.V7Actor.CreateDockerPackageByApplication(pushPlan.Application.GUID, pushPlan.DockerImageCredentials)
		if err != nil {
			return pushPlan, Warnings(warnings), err
		}

		eventStream <- SetDockerImageComplete

		polledPackage, pollWarnings, err := actor.V7Actor.PollPackage(pkg)

		pushPlan.PackageGUID = polledPackage.GUID

		return pushPlan, Warnings(append(warnings, pollWarnings...)), err
	}

	return pushPlan, nil, nil
}
