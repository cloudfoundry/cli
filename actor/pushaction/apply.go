package pushaction

import log "github.com/Sirupsen/logrus"

func (actor Actor) Apply(config ApplicationConfig) (<-chan Event, <-chan Warnings, <-chan error) {
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting apply go routine")
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		if config.CurrentApplication.GUID != "" {
			log.Debugf("updating application: %#v", config.DesiredApplication)
			_, warnings, err := actor.V2Actor.UpdateApplication(config.DesiredApplication)
			warningsStream <- Warnings(warnings)
			if err != nil {
				log.Errorln("updating application:", err)
				errorStream <- err
				return
			}
			eventStream <- ApplicationUpdated
		} else {
			log.Debugf("creating application: %#v", config.DesiredApplication)
			_, warnings, err := actor.V2Actor.CreateApplication(config.DesiredApplication)
			warningsStream <- Warnings(warnings)
			if err != nil {
				log.Errorln("creating application:", err)
				errorStream <- err
				return
			}
			eventStream <- ApplicationCreated
		}

		log.Debug("completed apply")
		eventStream <- Complete
	}()

	return eventStream, warningsStream, errorStream
}
