package pushaction

import (
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/Sirupsen/logrus"
)

type ApplicationConfig struct {
	CurrentApplication v2action.Application
	DesiredApplication v2action.Application

	TargetedSpaceGUID string
	Path              string
}

func (actor Actor) Apply(config ApplicationConfig) (<-chan Event, <-chan Warnings, <-chan error) {
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting apply thread")
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		if config.CurrentApplication.GUID != "" {
			log.Debugf("updating application: %#v", config.DesiredApplication)
			_, warnings, err := actor.V2Actor.UpdateApplication(config.DesiredApplication)
			warningsStream <- Warnings(warnings)
			if err != nil {
				log.Errorf("error updating application: %v", err)
				errorStream <- err
				return
			}
			eventStream <- ApplicationUpdated
		} else {
			log.Debugf("creating application: %#v", config.DesiredApplication)
			_, warnings, err := actor.V2Actor.CreateApplication(config.DesiredApplication)
			warningsStream <- Warnings(warnings)
			if err != nil {
				log.Errorf("error creating application: %v", err)
				errorStream <- err
				return
			}
			eventStream <- ApplicationCreated
		}

		log.Debug("sending complete")
		eventStream <- Complete
		log.Debug("sent complete")
	}()

	return eventStream, warningsStream, errorStream
}

func (actor Actor) ConvertToApplicationConfig(spaceGUID string, apps []manifest.Application) ([]ApplicationConfig, Warnings, error) {
	var configs []ApplicationConfig
	var warnings Warnings

	log.Infof("iterating through %d app configuration(s)", len(apps))
	for _, app := range apps {
		log.Infof("searching for app %s", app.Name)

		foundApp, v2Warnings, err := actor.V2Actor.GetApplicationByNameAndSpace(app.Name, spaceGUID)
		warnings = append(warnings, v2Warnings...)
		if _, ok := err.(v2action.ApplicationNotFoundError); ok {
			log.Warnf("unable to find app %s in current space (GUID: %s)", app.Name, spaceGUID)
		} else if err != nil {
			log.Errorf("error finding app %s", app.Name)
			return nil, warnings, err
		}

		config := ApplicationConfig{
			CurrentApplication: foundApp,
			DesiredApplication: foundApp,
			TargetedSpaceGUID:  spaceGUID,
			Path:               app.Path,
		}

		config.DesiredApplication.Name = app.Name
		config.DesiredApplication.SpaceGUID = spaceGUID
		configs = append(configs, config)
	}
	return configs, warnings, nil
}
