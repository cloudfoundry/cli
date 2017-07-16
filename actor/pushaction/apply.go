package pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	log "github.com/sirupsen/logrus"
)

const PushRetries = 3

type UploadFailedError struct {
	Err error
}

func (UploadFailedError) Error() string {
	return "upload failed"
}

func (actor Actor) Apply(config ApplicationConfig, progressBar ProgressBar) (<-chan ApplicationConfig, <-chan Event, <-chan Warnings, <-chan error) {
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

		eventStream <- SettingUpApplication
		config, event, warnings, err = actor.CreateOrUpdateApp(config)
		warningsStream <- warnings
		if err != nil {
			errorStream <- err
			return
		}
		eventStream <- event
		log.Debugf("desired application: %#v", config.DesiredApplication)

		eventStream <- ConfiguringRoutes

		var createdRoutes bool
		config, createdRoutes, warnings, err = actor.CreateRoutes(config)
		warningsStream <- warnings
		if err != nil {
			errorStream <- err
			return
		}
		if createdRoutes {
			log.Debugf("updated desired routes: %#v", config.DesiredRoutes)
			eventStream <- CreatedRoutes
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
			eventStream <- BoundRoutes
		}

		if config.DesiredApplication.DockerImage == "" {
			eventStream <- ResourceMatching
			config, warnings = actor.SetMatchedResources(config)
			warningsStream <- warnings

			archivePath, err := actor.CreateArchive(config)
			if err != nil {
				errorStream <- err
				return
			}
			eventStream <- CreatingArchive
			defer os.Remove(archivePath)

			for count := 0; count < PushRetries; count++ {
				warnings, err = actor.UploadPackage(config, archivePath, progressBar, eventStream)
				warningsStream <- warnings
				if _, ok := err.(ccerror.PipeSeekError); !ok {
					break
				}
				eventStream <- RetryUpload
			}

			if err != nil {
				if _, ok := err.(ccerror.PipeSeekError); ok {
					errorStream <- UploadFailedError{}
					return
				}
				errorStream <- err
				return
			}
		} else {
			log.WithField("docker_image", config.DesiredApplication.DockerImage).Debug("skipping file upload")
		}

		configStream <- config

		log.Debug("completed apply")
		eventStream <- Complete
	}()

	return configStream, eventStream, warningsStream, errorStream
}
