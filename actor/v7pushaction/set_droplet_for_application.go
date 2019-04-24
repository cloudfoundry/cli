package v7pushaction

func (actor Actor) SetDropletForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- SettingDroplet

	warnings, err := actor.V7Actor.SetApplicationDroplet(pushPlan.Application.GUID, pushPlan.DropletGUID)
	if err != nil {
		return pushPlan, Warnings(warnings), err
	}

	eventStream <- SetDropletComplete

	return pushPlan, Warnings(warnings), nil
}
