package v7pushaction

import (
	log "github.com/sirupsen/logrus"
)

func (actor Actor) Actualize(plan PushPlan, progressBar ProgressBar) (
	<-chan PushPlan, <-chan Event, <-chan Warnings, <-chan error,
) {
	log.Debugln("Starting to Actualize Push plan:", plan)
	planStream := make(chan PushPlan)
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting actualize go routine")
		defer close(planStream)
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		changeFuncs := actor.ChangeApplicationFuncs

		if plan.NoStart {
			changeFuncs = append(changeFuncs, actor.NoStartFuncs...)
		} else {
			changeFuncs = append(changeFuncs, actor.StartFuncs...)
		}

		var err error
		var warnings Warnings
		for _, changeAppFunc := range changeFuncs {
			plan, warnings, err = changeAppFunc(plan, eventStream, progressBar)
			warningsStream <- warnings
			if err != nil {
				errorStream <- err
				return
			}
			planStream <- plan
		}

		log.Debug("completed apply")
		eventStream <- Complete
	}()
	return planStream, eventStream, warningsStream, errorStream
}
