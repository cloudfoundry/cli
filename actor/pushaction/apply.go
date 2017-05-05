package pushaction

import log "github.com/Sirupsen/logrus"

func (actor Actor) Apply(config ApplicationConfig) (<-chan ApplicationConfig, <-chan Event, <-chan Warnings, <-chan error) {
	configStream := make(chan ApplicationConfig)
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting apply go routine")
		defer close(configStream)
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		var event Event
		var warnings Warnings
		var err error
		config, event, warnings, err = actor.CreateOrUpdateApp(config)
		warningsStream <- warnings
		if err != nil {
			errorStream <- err
			return
		}
		eventStream <- event
		log.Debugf("desired application: %#v", config.DesiredApplication)

		var createdRoutes bool
		config, createdRoutes, warnings, err = actor.CreateRoutes(config)
		warningsStream <- warnings
		if err != nil {
			errorStream <- err
			return
		}
		if createdRoutes {
			log.Debugf("updated desired routes: %#v", config.DesiredRoutes)
			eventStream <- RouteCreated
		}

		var boundRoutes bool
		config, boundRoutes, warnings, err = actor.BindRoutes(config)
		warningsStream <- warnings
		if err != nil {
			errorStream <- err
			return
		}
		if boundRoutes {
			log.Debugf("updated desired routes: %#v", config.DesiredRoutes)
			eventStream <- RouteBound
		}

		archivePath, err := actor.CreateArchive(config)
		if err != nil {
			errorStream <- err
			return
		}

		warnings, err = actor.UploadPackage(config, archivePath, eventStream)
		warningsStream <- warnings
		if err != nil {
			errorStream <- err
			return
		}

		configStream <- config

		log.Debug("completed apply")
		eventStream <- Complete
	}()

	return configStream, eventStream, warningsStream, errorStream
}
