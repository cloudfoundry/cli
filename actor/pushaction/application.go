package pushaction

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/sirupsen/logrus"
)

type Application struct {
	v2action.Application
	Stack v2action.Stack
}

func (actor Actor) CreateOrUpdateApp(config ApplicationConfig) (ApplicationConfig, Event, Warnings, error) {
	log.Debugf("creating or updating application")
	if config.UpdatingApplication() {
		log.Debugf("updating application: %#v", config.DesiredApplication)
		app, warnings, err := actor.V2Actor.UpdateApplication(config.DesiredApplication.Application)
		if err != nil {
			log.Errorln("updating application:", err)
			return ApplicationConfig{}, "", Warnings(warnings), err
		}

		config.DesiredApplication.Application = app
		config.CurrentApplication = config.DesiredApplication
		return config, UpdatedApplication, Warnings(warnings), err
	} else {
		log.Debugf("creating application: %#v", config.DesiredApplication)
		app, warnings, err := actor.V2Actor.CreateApplication(config.DesiredApplication.Application)
		if err != nil {
			log.Errorln("creating application:", err)
			return ApplicationConfig{}, "", Warnings(warnings), err
		}

		config.DesiredApplication.Application = app
		config.CurrentApplication = config.DesiredApplication
		return config, CreatedApplication, Warnings(warnings), err
	}
}

func (actor Actor) FindOrReturnPartialApp(appName string, spaceGUID string) (bool, Application, v2action.Warnings, error) {
	foundApp, v2Warnings, err := actor.V2Actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if _, ok := err.(v2action.ApplicationNotFoundError); ok {
		log.Warnf("unable to find app %s in current space (GUID: %s)", appName, spaceGUID)
		return false, Application{}, v2Warnings, nil
	}

	app := Application{
		Application: foundApp,
	}
	return true, app, v2Warnings, err
}
