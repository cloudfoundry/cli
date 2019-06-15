package v7pushaction

import (
	log "github.com/sirupsen/logrus"
)

func (actor Actor) Actualize(plan PushPlan, progressBar ProgressBar) <-chan *PushEvent {
	log.Debugln("Starting to Actualize Push plan:", plan)
	eventStream := make(chan *PushEvent)

	go func() {
		log.Debug("starting actualize go routine")
		defer close(eventStream)

		var err error
		var warnings Warnings
		for _, changeAppFunc := range actor.ChangeApplicationSequence(plan) {
			plan, warnings, err = changeAppFunc(plan, eventStream, progressBar)
			eventStream <- &PushEvent{Plan: plan, Err: err, Warnings: warnings}
			if err != nil {
				return
			}
		}

		log.Debug("completed apply")
	}()

	return eventStream
}
