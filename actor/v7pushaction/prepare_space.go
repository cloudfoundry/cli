package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/util/manifestparser"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) PrepareSpace(spaceGUID string, appName string, manifestParser *manifestparser.Parser, overrides FlagOverrides) (<-chan []string, <-chan Event, <-chan Warnings, <-chan error) {
	log.Debugln("Starting to Actualize Space:", spaceGUID)
	appNameStream := make(chan []string)
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting space preparation go routine")
		defer close(appNameStream)
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		var warnings v7action.Warnings
		var appNames []string
		var err error
		var successEvent Event

		if manifestParser.FullRawManifest() == nil {
			app := v7action.Application{Name: appName}
			if overrides.DockerImage != "" {
				app.LifecycleType = constant.AppLifecycleTypeDocker
			}
			_, warnings, err = actor.V7Actor.CreateApplicationInSpace(app, spaceGUID)
			if _, ok := err.(actionerror.ApplicationAlreadyExistsError); ok {
				eventStream <- SkippingApplicationCreation
				successEvent = ApplicationAlreadyExists
				err = nil
			} else {
				eventStream <- CreatingApplication
				successEvent = CreatedApplication
			}
		} else {
			var manifest []byte
			manifest, err = getManifest(appName, manifestParser)
			if err != nil {
				errorStream <- err
				return
			}
			eventStream <- ApplyManifest
			warnings, err = actor.V7Actor.SetSpaceManifest(spaceGUID, manifest)
			successEvent = ApplyManifestComplete
		}

		if appName != "" {
			appNames = []string{appName}
		} else {
			appNames = manifestParser.AppNames()
		}

		warningsStream <- Warnings(warnings)
		errorStream <- err
		if err != nil {
			return
		}
		appNameStream <- appNames
		eventStream <- successEvent
	}()

	log.WithFields(log.Fields{"appName": appName, "spaceGUID": spaceGUID}).Info("Application Look Up")

	return appNameStream, eventStream, warningsStream, errorStream
}

func getManifest(appName string, parser *manifestparser.Parser) ([]byte, error) {
	if appName != "" {
		return parser.RawAppManifest(appName)
	}
	return parser.FullRawManifest(), nil
}
