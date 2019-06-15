package v7pushaction

func (actor Actor) SetDropletForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Plan: pushPlan, Event: SettingDroplet}

	warnings, err := actor.V7Actor.SetApplicationDroplet(pushPlan.Application.GUID, pushPlan.DropletGUID)
	if err != nil {
		return pushPlan, Warnings(warnings), err
	}

	eventStream <- &PushEvent{Plan: pushPlan, Event: SetDropletComplete}

	return pushPlan, Warnings(warnings), nil
}
