package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) PrepareSpace(pushPlans []PushPlan, manifestParser ManifestParser) (<-chan []PushPlan, <-chan Event, <-chan Warnings, <-chan error) {
	pushPlansStream := make(chan []PushPlan)
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting space preparation go routine")
		defer close(pushPlansStream)
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		var warnings v7action.Warnings
		var err error
		var successEvent Event

		if manifestParser.ContainsManifest() {
			var manifest []byte
			manifest, err = getManifest(pushPlans, manifestParser)
			if err != nil {
				errorStream <- err
				return
			}
			eventStream <- ApplyManifest
			warnings, err = actor.V7Actor.SetSpaceManifest(pushPlans[0].SpaceGUID, manifest)
			successEvent = ApplyManifestComplete
		} else {
			_, warnings, err = actor.V7Actor.CreateApplicationInSpace(pushPlans[0].Application, pushPlans[0].SpaceGUID)
			if _, ok := err.(actionerror.ApplicationAlreadyExistsError); ok {
				eventStream <- SkippingApplicationCreation
				successEvent = ApplicationAlreadyExists
				err = nil
			} else {
				eventStream <- CreatingApplication
				successEvent = CreatedApplication
			}
		}

		warningsStream <- Warnings(warnings)
		errorStream <- err
		if err != nil {
			return
		}
		pushPlansStream <- pushPlans
		eventStream <- successEvent
	}()

	return pushPlansStream, eventStream, warningsStream, errorStream
}

func getManifest(plans []PushPlan, parser ManifestParser) ([]byte, error) {
	if len(plans) == 1 {
		return parser.RawAppManifest(plans[0].Application.Name)
	}
	return parser.FullRawManifest(), nil
}
