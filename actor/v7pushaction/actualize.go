package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	log "github.com/sirupsen/logrus"
)

const PushRetries = 3

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

		var err error
		var wrgs Warnings
		for _, changeAppFunc := range actor.ChangeApplicationFuncs {
			plan, wrgs, err = changeAppFunc(plan, eventStream, progressBar)
			// (NEW) events can happen here now
			warningsStream <- wrgs
			if err != nil {
				errorStream <- err
				return
			}
			// (OLD) used to happen here
			planStream <- plan
		}

		var warnings v7action.Warnings
		if plan.NoStart {
			if plan.Application.State == constant.ApplicationStarted {
				eventStream <- StoppingApplication
				warnings, err = actor.V7Actor.StopApplication(plan.Application.GUID)
				warningsStream <- Warnings(warnings)
				if err != nil {
					errorStream <- err
				}
				eventStream <- StoppingApplicationComplete
			}
			eventStream <- Complete
			return
		}

		eventStream <- StartingStaging

		build, warnings, err := actor.V7Actor.StageApplicationPackage(plan.PackageGUID)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- PollingBuild

		droplet, warnings, err := actor.V7Actor.PollBuild(build.GUID, plan.Application.Name)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- StagingComplete
		eventStream <- SettingDroplet

		warnings, err = actor.V7Actor.SetApplicationDroplet(plan.Application.GUID, droplet.GUID)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- SetDropletComplete

		log.Debug("completed apply")
		eventStream <- Complete
	}()
	return planStream, eventStream, warningsStream, errorStream
}
