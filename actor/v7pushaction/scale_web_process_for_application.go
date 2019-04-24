package v7pushaction

import (
	log "github.com/sirupsen/logrus"
)

func (actor Actor) ScaleWebProcessForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	if pushPlan.ScaleWebProcessNeedsUpdate {
		log.Info("Scaling Web Process")
		eventStream <- ScaleWebProcess

		warnings, err := actor.V7Actor.ScaleProcessByApplication(pushPlan.Application.GUID, pushPlan.ScaleWebProcess)

		if err != nil {
			return pushPlan, Warnings(warnings), err
		}
		eventStream <- ScaleWebProcessComplete
		return pushPlan, Warnings(warnings), nil
	}

	return pushPlan, nil, nil
}
