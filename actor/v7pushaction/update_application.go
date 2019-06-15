package v7pushaction

import (
	log "github.com/sirupsen/logrus"
)

func (actor Actor) UpdateApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	log.WithField("Name", pushPlan.Application.Name).Info("updating app")

	application, warnings, err := actor.V7Actor.UpdateApplication(pushPlan.Application)
	pushPlan.Application = application

	return pushPlan, Warnings(warnings), err
}
