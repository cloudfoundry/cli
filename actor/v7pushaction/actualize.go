package v7pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	log "github.com/sirupsen/logrus"
)

const PushRetries = 3

func (actor Actor) Actualize(state PushState, progressBar ProgressBar) (
	<-chan PushState, <-chan Event, <-chan Warnings, <-chan error,
) {
	log.Debugln("Starting to Actualize Push State:", state)
	stateStream := make(chan PushState)
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting actualize go routine")
		defer close(stateStream)
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		if state.Application.GUID != "" {
			log.WithField("Name", state.Application.Name).Info("skipping app creation as it has a GUID")
			eventStream <- SkippingApplicationCreation
			application, warnings, err := actor.V7Actor.UpdateApplication(state.Application)
			state.Application = application
			warningsStream <- Warnings(warnings)
			if err != nil {
				errorStream <- err
				return
			}
			eventStream <- UpdatedApplication

		} else {
			log.WithField("Name", state.Application.Name).Info("creating app")
			createdApp, createWarnings, err := actor.V7Actor.CreateApplicationInSpace(state.Application, state.SpaceGUID)
			warningsStream <- Warnings(createWarnings)
			if err != nil {
				errorStream <- err
				return
			}

			state.Application = createdApp
			eventStream <- CreatedApplication
		}
		stateStream <- state

		if state.Overrides.Memory.IsSet {
			log.Info("Scaling Web Process")
			eventStream <- ScaleWebProcess
			process := v7action.Process{Type: constant.ProcessTypeWeb, MemoryInMB: state.Overrides.Memory}
			scaleWarnings, err := actor.V7Actor.ScaleProcessByApplication(state.Application.GUID, process)
			warningsStream <- Warnings(scaleWarnings)
			if err != nil {
				errorStream <- err
				return
			}
			eventStream <- ScaleWebProcessComplete
		}

		if state.Overrides.StartCommand.IsSet || state.Overrides.HealthCheckType != "" {
			log.Info("Setting Web Process's Configuration")
			eventStream <- SetProcessConfiguration

			var process v7action.Process
			if state.Overrides.StartCommand.IsSet {
				process.Command = state.Overrides.StartCommand.Value
			}
			if state.Overrides.HealthCheckType != "" {
				process.HealthCheckType = state.Overrides.HealthCheckType
				process.HealthCheckEndpoint = constant.ProcessHealthCheckEndpointDefault
			}

			log.WithField("Process", process).Debug("Update process")
			warnings, err := actor.V7Actor.UpdateProcessByTypeAndApplication(constant.ProcessTypeWeb, state.Application.GUID, process)
			warningsStream <- Warnings(warnings)
			if err != nil {
				errorStream <- err
				return
			}
			eventStream <- SetProcessConfigurationComplete
		}

		eventStream <- CreatingAndMappingRoutes
		routeWarnings, routeErr := actor.CreateAndMapDefaultApplicationRoute(state.OrgGUID, state.SpaceGUID, state.Application)
		warningsStream <- Warnings(routeWarnings)
		if routeErr != nil {
			errorStream <- routeErr
			return
		}
		eventStream <- CreatedRoutes

		log.WithField("Path", state.BitsPath).Info(string(CreatingArchive))

		eventStream <- CreatingArchive
		archivePath, err := actor.SharedActor.ZipDirectoryResources(state.BitsPath, state.AllResources)
		if err != nil {
			errorStream <- err
			return
		}
		defer os.RemoveAll(archivePath)

		eventStream <- CreatingPackage
		log.WithField("GUID", state.Application.GUID).Info("creating package")
		pkg, warnings, err := actor.V7Actor.CreateBitsPackageByApplication(state.Application.GUID)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		for count := 0; count < PushRetries; count++ {
			eventStream <- ReadingArchive
			log.WithField("GUID", state.Application.GUID).Info("creating package")
			file, size, readErr := actor.SharedActor.ReadArchive(archivePath)
			if readErr != nil {
				errorStream <- readErr
				return
			}
			defer file.Close()

			eventStream <- UploadingApplicationWithArchive
			progressReader := progressBar.NewProgressBarWrapper(file, size)
			pkg, warnings, err = actor.V7Actor.UploadBitsPackage(pkg, state.MatchedResources, progressReader, size)
			warningsStream <- Warnings(warnings)

			if _, ok := err.(ccerror.PipeSeekError); ok {
				eventStream <- RetryUpload
				continue
			}
			break
		}

		if err != nil {
			if e, ok := err.(ccerror.PipeSeekError); ok {
				errorStream <- actionerror.UploadFailedError{Err: e.Err}
				return
			}
			errorStream <- err
			return
		}

		eventStream <- UploadWithArchiveComplete

		polledPackage, warnings, err := actor.V7Actor.PollPackage(pkg)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- StartingStaging

		build, warnings, err := actor.V7Actor.StageApplicationPackage(polledPackage.GUID)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- PollingBuild

		droplet, warnings, err := actor.V7Actor.PollBuild(build.GUID, state.Application.Name)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- StagingComplete
		eventStream <- SettingDroplet

		warnings, err = actor.V7Actor.SetApplicationDroplet(state.Application.GUID, droplet.GUID)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- SetDropletComplete

		log.Debug("completed apply")
		eventStream <- Complete
	}()
	return stateStream, eventStream, warningsStream, errorStream
}
