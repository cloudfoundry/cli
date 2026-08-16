package v7pushaction

import (
	"log/slog"
)

func (actor Actor) Actualize(plan PushPlan, progressBar ProgressBar) <-chan *PushEvent {
	slog.Debug("Starting to Actualize Push plan", "plan", plan)
	eventStream := make(chan *PushEvent)

	go func() {
		slog.Debug("starting actualize go routine")
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

		slog.Debug("completed apply")
	}()

	return eventStream
}
