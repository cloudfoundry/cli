package pushaction

import log "github.com/sirupsen/logrus"

func (actor Actor) CreateOrUpdateApp(config ApplicationConfig) (ApplicationConfig, Event, Warnings, error) {
	log.Debugf("creating or updating application")
	if config.UpdatingApplication() {
		log.Debugf("updating application: %#v", config.DesiredApplication)
		app, warnings, err := actor.V2Actor.UpdateApplication(config.DesiredApplication)
		if err != nil {
			log.Errorln("updating application:", err)
			return ApplicationConfig{}, "", Warnings(warnings), err
		}

		config.DesiredApplication = app
		config.CurrentApplication = config.DesiredApplication
		return config, UpdatedApplication, Warnings(warnings), err
	} else {
		log.Debugf("creating application: %#v", config.DesiredApplication)
		app, warnings, err := actor.V2Actor.CreateApplication(config.DesiredApplication)
		if err != nil {
			log.Errorln("creating application:", err)
			return ApplicationConfig{}, "", Warnings(warnings), err
		}

		config.DesiredApplication = app
		config.CurrentApplication = config.DesiredApplication
		return config, CreatedApplication, Warnings(warnings), err
	}
}
