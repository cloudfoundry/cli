package v7pushaction

import log "github.com/sirupsen/logrus"

func (actor Actor) RestartApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	log.Info("Restarting Application")
	var allWarnings Warnings
	eventStream <- &PushEvent{Plan: pushPlan, Event: RestartingApplication}
	warnings, err := actor.V7Actor.RestartApplication(pushPlan.Application.GUID, pushPlan.NoWait)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return pushPlan, allWarnings, err
	}
	warnings, err = actor.V7Actor.PollStart(pushPlan.Application.GUID, pushPlan.NoWait)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return pushPlan, allWarnings, err
	}
	eventStream <- &PushEvent{Plan: pushPlan, Event: RestartingApplicationComplete}

	return pushPlan, allWarnings, nil
}
