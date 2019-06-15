package v7pushaction

import (
	log "github.com/sirupsen/logrus"
)

func (actor Actor) ScaleWebProcessForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	if pushPlan.ScaleWebProcessNeedsUpdate {
		log.Info("Scaling Web Process")
		eventStream <- &PushEvent{Event: ScaleWebProcess}

		warnings, err := actor.V7Actor.ScaleProcessByApplication(pushPlan.Application.GUID, pushPlan.ScaleWebProcess)

		if err != nil {
			return pushPlan, Warnings(warnings), err
		}
		eventStream <- &PushEvent{Event: ScaleWebProcessComplete}
		return pushPlan, Warnings(warnings), nil
	}

	return pushPlan, nil, nil
}
