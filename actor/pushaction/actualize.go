package pushaction

import (
	log "github.com/sirupsen/logrus"
)

func (actor Actor) Actualize(state PushState, progressBar ProgressBar) (
	<-chan PushState, <-chan Event, <-chan Warnings, <-chan error,
) {
	stateStream := make(chan PushState)
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting actualize go routine")
		defer close(stateStream)
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		if state.Application.GUID != "" {
			log.WithField("Name", state.Application.Name).Info("skipping app creation as it has a GUID")
			eventStream <- SkipingApplicationCreation
		} else {
			log.WithField("Name", state.Application.Name).Info("creating app")
			createdApp, createWarnings, err := actor.V3Actor.CreateApplicationInSpace(state.Application, state.SpaceGUID)
			warningsStream <- Warnings(createWarnings)
			if err != nil {
				errorStream <- err
				return
			}

			state.Application = createdApp
			eventStream <- CreatedApplication
		}

		stateStream <- state

		log.Debug("completed apply")
		eventStream <- Complete
	}()
	return stateStream, eventStream, warningsStream, errorStream
}
