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

		var err error

		state, err = actor.UpdateApplication(state, warningsStream, eventStream)
		if err != nil {
			errorStream <- err
			return
		}
		stateStream <- state

		// TODO Remove this
		if len(state.Manifest) > 0 {
			err = actor.ApplyManifest(state, warningsStream, eventStream)
			if err != nil {
				errorStream <- err
				return
			}
		} else if !state.Overrides.SkipRouteCreation {
			eventStream <- CreatingAndMappingRoutes
			routeWarnings, routeErr := actor.CreateAndMapDefaultApplicationRoute(state.OrgGUID, state.SpaceGUID, state.Application)
			warningsStream <- Warnings(routeWarnings)
			if routeErr != nil {
				errorStream <- routeErr
				return
			}
			eventStream <- CreatedRoutes
		}

		err = actor.ScaleProcess(state, warningsStream, eventStream)
		if err != nil {
			errorStream <- err
			return
		}

		err = actor.UpdateProcess(state, warningsStream, eventStream)
		if err != nil {
			errorStream <- err
			return
		}

		pkg, err := actor.CreatePackage(state, progressBar, warningsStream, eventStream)
		if err != nil {
			errorStream <- err
			return
		}

		polledPackage, warnings, err := actor.V7Actor.PollPackage(pkg)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		if state.Overrides.NoStart == true {
			if state.Application.State == constant.ApplicationStarted {
				eventStream <- StoppingApplication
				warnings, err = actor.V7Actor.StopApplication(state.Application.GUID)
				warningsStream <- Warnings(warnings)
				if err != nil {
					errorStream <- err
				}
				eventStream <- StoppingApplicationComplete
			}
			eventStream <- Complete
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

func (actor Actor) ApplyManifest(state PushState, warningsStream chan Warnings, eventStream chan Event) error {
	eventStream <- ApplyManifest
	warnings, err := actor.V7Actor.SetApplicationManifest(state.Application.GUID, state.Manifest)
	warningsStream <- Warnings(warnings)
	if err != nil {
		return err
	}
	eventStream <- ApplyManifestComplete

	return nil
}

func (actor Actor) CreateAndUploadApplicationBits(state PushState, progressBar ProgressBar, warningsStream chan Warnings, eventStream chan Event) (v7action.Package, error) {
	log.WithField("Path", state.BitsPath).Info(string(CreatingArchive))

	eventStream <- CreatingArchive
	archivePath, err := actor.GetArchivePath(state)
	if err != nil {
		return v7action.Package{}, err
	}
	defer os.RemoveAll(archivePath)

	eventStream <- CreatingPackage
	log.WithField("GUID", state.Application.GUID).Info("creating package")
	pkg, warnings, err := actor.V7Actor.CreateBitsPackageByApplication(state.Application.GUID)
	warningsStream <- Warnings(warnings)
	if err != nil {
		return v7action.Package{}, err
	}

	// Uploading package/app bits
	for count := 0; count < PushRetries; count++ {
		eventStream <- ReadingArchive
		log.WithField("GUID", state.Application.GUID).Info("reading archive")
		file, size, readErr := actor.SharedActor.ReadArchive(archivePath)
		if readErr != nil {
			return v7action.Package{}, readErr
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
			return v7action.Package{}, actionerror.UploadFailedError{Err: e.Err}
		}
		return v7action.Package{}, err
	}

	eventStream <- UploadWithArchiveComplete
	return pkg, nil
}

func (actor Actor) UpdateApplication(state PushState, warningsStream chan Warnings, eventStream chan Event) (PushState, error) {
	if !state.ApplicationNeedsUpdate {
		return state, nil
	}

	log.WithField("Name", state.Application.Name).Info("updating app")
	eventStream <- SkippingApplicationCreation

	application, warnings, err := actor.V7Actor.UpdateApplication(state.Application)
	state.Application = application
	warningsStream <- Warnings(warnings)
	if err != nil {
		return state, err
	}

	eventStream <- UpdatedApplication

	return state, nil
}

func (actor Actor) CreatePackage(state PushState, progressBar ProgressBar, warningsStream chan Warnings, eventStream chan Event) (v7action.Package, error) {
	if state.Application.LifecycleType == constant.AppLifecycleTypeDocker {
		eventStream <- SetDockerImage
		pkg, warnings, err := actor.V7Actor.CreateDockerPackageByApplication(state.Application.GUID, v7action.DockerImageCredentials{
			Path:     state.Overrides.DockerImage,
			Username: state.Overrides.DockerUsername,
			Password: state.Overrides.DockerPassword,
		})
		warningsStream <- Warnings(warnings)
		if err != nil {
			return v7action.Package{}, err
		}
		eventStream <- SetDockerImageComplete
		return pkg, nil
	}

	return actor.CreateAndUploadApplicationBits(state, progressBar, warningsStream, eventStream)
}

func (actor Actor) GetArchivePath(state PushState) (string, error) {
	if state.Archive {
		return actor.SharedActor.ZipArchiveResources(state.BitsPath, state.AllResources)
	}
	return actor.SharedActor.ZipDirectoryResources(state.BitsPath, state.AllResources)
}

func (actor Actor) ScaleProcess(state PushState, warningsStream chan Warnings, eventStream chan Event) error {
	if shouldScaleProcess(state) {
		log.Info("Scaling Web Process")
		eventStream <- ScaleWebProcess

		process := v7action.Process{
			Type:       constant.ProcessTypeWeb,
			MemoryInMB: state.Overrides.Memory,
			DiskInMB:   state.Overrides.Disk,
			Instances:  state.Overrides.Instances,
		}
		scaleWarnings, err := actor.V7Actor.ScaleProcessByApplication(state.Application.GUID, process)
		warningsStream <- Warnings(scaleWarnings)
		if err != nil {
			return err
		}
		eventStream <- ScaleWebProcessComplete
	}

	return nil
}

func shouldScaleProcess(state PushState) bool {
	return state.Overrides.Memory.IsSet || state.Overrides.Instances.IsSet || state.Overrides.Disk.IsSet
}

func (actor Actor) UpdateProcess(state PushState, warningsStream chan Warnings, eventStream chan Event) error {
	if state.Overrides.StartCommand.IsSet || state.Overrides.HealthCheckType != "" {
		log.Info("Setting Web Process's Configuration")
		eventStream <- SetProcessConfiguration

		var process v7action.Process
		if state.Overrides.StartCommand.IsSet {
			process.Command = state.Overrides.StartCommand
		}
		if state.Overrides.HealthCheckType != "" {
			process.HealthCheckType = state.Overrides.HealthCheckType
			process.HealthCheckEndpoint = state.Overrides.HealthCheckEndpoint
		}

		log.WithField("Process", process).Debug("Update process")
		warnings, err := actor.V7Actor.UpdateProcessByTypeAndApplication(constant.ProcessTypeWeb, state.Application.GUID, process)
		warningsStream <- Warnings(warnings)
		if err != nil {
			return err
		}
		eventStream <- SetProcessConfigurationComplete
	}

	return nil
}
