package pushaction

import (
	"code.cloudfoundry.org/cli/actor/v2action"
	log "github.com/sirupsen/logrus"
)

type Application struct {
	v2action.Application
	Stack v2action.Stack
}

func (app *Application) SetStack(stack v2action.Stack) {
	app.Stack = stack
	app.StackGUID = stack.GUID
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
	foundApp, appWarnings, err := actor.V2Actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if _, ok := err.(v2action.ApplicationNotFoundError); ok {
		log.Warnf("unable to find app %s in current space (GUID: %s)", appName, spaceGUID)
		return false, Application{
			Application: v2action.Application{
				Name:      appName,
				SpaceGUID: spaceGUID,
			},
		}, appWarnings, nil
	} else if err != nil {
		log.WithField("appName", appName).Error("error retrieving app")
		return false, Application{}, appWarnings, err
	}

	stack, stackWarnings, err := actor.V2Actor.GetStack(foundApp.StackGUID)
	warnings := append(appWarnings, stackWarnings...)
	if err != nil {
		log.Warnf("unable to find app's stack (GUID: %s)", foundApp.StackGUID)
		return false, Application{}, warnings, err
	}

	app := Application{
		Application: foundApp,
		Stack:       stack,
	}
	return true, app, warnings, err
}
