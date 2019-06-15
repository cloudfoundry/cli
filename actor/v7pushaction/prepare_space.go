package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) PrepareSpace(pushPlans []PushPlan, manifestParser ManifestParser) ([]string, <-chan *PushEvent) {
	pushEventStream := make(chan *PushEvent)

	go func() {
		log.Debug("starting apply manifest go routine")
		defer close(pushEventStream)

		var warnings v7action.Warnings
		var err error
		var successEvent Event

		if manifestParser.ContainsManifest() {
			var manifest []byte
			manifest, err = getManifest(pushPlans, manifestParser)
			if err != nil {
				pushEventStream <- &PushEvent{Err: err}
				return
			}
			pushEventStream <- &PushEvent{Event: ApplyManifest}
			warnings, err = actor.V7Actor.SetSpaceManifest(pushPlans[0].SpaceGUID, manifest, pushPlans[0].NoRouteFlag)
			successEvent = ApplyManifestComplete
		} else {
			_, warnings, err = actor.V7Actor.CreateApplicationInSpace(pushPlans[0].Application, pushPlans[0].SpaceGUID)
			if _, ok := err.(actionerror.ApplicationAlreadyExistsError); ok {
				pushEventStream <- &PushEvent{Event: SkippingApplicationCreation, Plan: pushPlans[0]}
				successEvent = ApplicationAlreadyExists
				err = nil
			} else {
				pushEventStream <- &PushEvent{Event: CreatingApplication, Plan: pushPlans[0]}
				successEvent = CreatedApplication
			}
		}

		if err != nil {
			pushEventStream <- &PushEvent{Err: err, Warnings: Warnings(warnings), Plan: pushPlans[0]}
			return
		}

		pushEventStream <- &PushEvent{Event: successEvent, Err: nil, Warnings: Warnings(warnings), Plan: pushPlans[0]}
	}()

	var appNames []string
	for _, plan := range pushPlans {
		appNames = append(appNames, plan.Application.Name)
	}

	return appNames, pushEventStream
}

func getManifest(plans []PushPlan, parser ManifestParser) ([]byte, error) {
	if len(plans) == 1 {
		return parser.RawAppManifest(plans[0].Application.Name)
	}
	return parser.FullRawManifest(), nil
}
