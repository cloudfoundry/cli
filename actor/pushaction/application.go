package pushaction

import log "github.com/Sirupsen/logrus"

func (actor Actor) CreateOrUpdateApp(config ApplicationConfig) (ApplicationConfig, Event, Warnings, error) {
	log.Debugf("creating or updating application")
	if config.DesiredApplication.GUID != "" {
		log.Debugf("updating application: %#v", config.DesiredApplication)
		app, warnings, err := actor.V2Actor.UpdateApplication(config.DesiredApplication)
		if err != nil {
			log.Errorln("updating application:", err)
			return ApplicationConfig{}, "", Warnings(warnings), err
		}

		config.DesiredApplication = app
		config.CurrentApplication = config.DesiredApplication
		return config, ApplicationUpdated, Warnings(warnings), err
	} else {
		log.Debugf("creating application: %#v", config.DesiredApplication)
		app, warnings, err := actor.V2Actor.CreateApplication(config.DesiredApplication)
		if err != nil {
			log.Errorln("creating application:", err)
			return ApplicationConfig{}, "", Warnings(warnings), err
		}

		config.DesiredApplication = app
		config.CurrentApplication = config.DesiredApplication
		return config, ApplicationCreated, Warnings(warnings), err
	}
}
