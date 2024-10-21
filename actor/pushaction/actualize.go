package pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccerror"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) Actualize(state PushPlan, progressBar ProgressBar) (
	<-chan PushPlan, <-chan Event, <-chan Warnings, <-chan error,
) {
	log.Debugf("Starting to Actualize Push plan: %#v\n", state)
	stateStream := make(chan PushPlan)
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
		} else {
			log.WithField("Name", state.Application.Name).Info("creating app")
			createdApp, createWarnings, err := actor.V3Actor.CreateApplicationInSpace(state.Application, state.SpaceGUID)
			warningsStream <- Warnings(createWarnings)
			if err != nil {
				errorStream <- err
				return
			}

			state.Application = createdApp
			eventStream <- CreatedApplication
		}
		stateStream <- state

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
		pkg, warnings, err := actor.V3Actor.CreateBitsPackageByApplication(state.Application.GUID)
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
			pkg, warnings, err = actor.V3Actor.UploadBitsPackage(pkg, state.MatchedResources, progressReader, size)
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

		polledPackage, warnings, err := actor.V3Actor.PollPackage(pkg)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- StartingStaging

		build, warnings, err := actor.V3Actor.StageApplicationPackage(polledPackage.GUID)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- PollingBuild

		droplet, warnings, err := actor.V3Actor.PollBuild(build.GUID, state.Application.Name)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- StagingComplete
		eventStream <- SettingDroplet

		warnings, err = actor.V3Actor.SetApplicationDroplet(state.Application.GUID, droplet.GUID)
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
