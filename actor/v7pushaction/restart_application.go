package v7pushaction

import log "github.com/sirupsen/logrus"

func (actor Actor) RestartApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	log.Info("Restarting Application")
	eventStream <- &PushEvent{Plan: pushPlan, Event: RestartingApplication}
	warnings, err := actor.V7Actor.RestartApplication(pushPlan.Application.GUID, pushPlan.NoWait)
	if err != nil {
		return pushPlan, Warnings(warnings), err
	}
	eventStream <- &PushEvent{Plan: pushPlan, Event: RestartingApplicationComplete}

	return pushPlan, Warnings(warnings), nil
}
