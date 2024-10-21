package pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccerror"

	log "github.com/sirupsen/logrus"
)

const PushRetries = 3

// Apply use the V2 API to create/update the application settings and
// eventually uploads the application bits.
//
// If multiple buildpacks are being applied to the application, the V3 API is
// used to set those buildpacks.
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
		if config.UpdatingApplication() {
			config, event, warnings, err = actor.UpdateApplication(config)
		} else {
			config, event, warnings, err = actor.CreateApplication(config)
		}

		warningsStream <- warnings
		if err != nil {
			errorStream <- err
			return
		}
		eventStream <- event
		log.Debugf("desired application: %#v", config.DesiredApplication)

		if config.NoRoute {
			if len(config.CurrentRoutes) > 0 {
				eventStream <- UnmappingRoutes
				config, warnings, err = actor.UnmapRoutes(config)
				warningsStream <- warnings
				if err != nil {
					errorStream <- err
					return
				}
			}
		} else {
			eventStream <- CreatingAndMappingRoutes

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
			config, boundRoutes, warnings, err = actor.MapRoutes(config)
			warningsStream <- warnings
			if err != nil {
				errorStream <- err
				return
			}
			if boundRoutes {
				log.Debugf("updated desired routes: %#v", config.DesiredRoutes)
				eventStream <- BoundRoutes
			}
		}

		if len(config.CurrentServices) != len(config.DesiredServices) {
			eventStream <- ConfiguringServices
			config, _, warnings, err = actor.BindServices(config)
			warningsStream <- warnings
			if err != nil {
				errorStream <- err
				return
			}

			log.Debugf("bound desired services: %#v", config.DesiredServices)
			eventStream <- BoundServices
		}

		if config.DropletPath != "" {
			for count := 0; count < PushRetries; count++ {
				warnings, err = actor.UploadDroplet(config, config.DropletPath, progressBar, eventStream)
				warningsStream <- warnings
				if _, ok := err.(ccerror.PipeSeekError); ok {
					eventStream <- RetryUpload
				} else {
					break
				}
			}

			if err != nil {
				if e, ok := err.(ccerror.PipeSeekError); ok {
					errorStream <- actionerror.UploadFailedError{Err: e.Err}
				} else {
					errorStream <- err
				}
				return
			}
		} else if config.DesiredApplication.DockerImage == "" {
			eventStream <- ResourceMatching
			config, warnings = actor.SetMatchedResources(config)
			warningsStream <- warnings

			if len(config.UnmatchedResources) > 0 {
				var archivePath string
				archivePath, err = actor.CreateArchive(config)
				if err != nil {
					errorStream <- err
					os.RemoveAll(archivePath)
					return
				}
				eventStream <- CreatingArchive
				defer os.RemoveAll(archivePath)

				for count := 0; count < PushRetries; count++ {
					warnings, err = actor.UploadPackageWithArchive(config, archivePath, progressBar, eventStream)
					warningsStream <- warnings
					if _, ok := err.(ccerror.PipeSeekError); ok {
						eventStream <- RetryUpload
					} else {
						break
					}
				}

				if err != nil {
					if e, ok := err.(ccerror.PipeSeekError); ok {
						errorStream <- actionerror.UploadFailedError{Err: e.Err}
					} else {
						errorStream <- err
					}
					return
				}
			} else {
				eventStream <- UploadingApplication
				warnings, err = actor.UploadPackage(config)
				warningsStream <- warnings
				if err != nil {
					errorStream <- err
					return
				}
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
