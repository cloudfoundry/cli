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

func (actor Actor) Actualize(plan PushPlan, progressBar ProgressBar) (
	<-chan PushPlan, <-chan Event, <-chan Warnings, <-chan error,
) {
	log.Debugln("Starting to Actualize Push plan:", plan)
	planStream := make(chan PushPlan)
	eventStream := make(chan Event)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		log.Debug("starting actualize go routine")
		defer close(planStream)
		defer close(eventStream)
		defer close(warningsStream)
		defer close(errorStream)

		var err error

		plan, err = actor.updateApplication(plan, warningsStream)
		if err != nil {
			errorStream <- err
			return
		}
		planStream <- plan

		if !plan.SkipRouteCreation {
			eventStream <- CreatingAndMappingRoutes
			routeWarnings, routeErr := actor.CreateAndMapDefaultApplicationRoute(plan.OrgGUID, plan.SpaceGUID, plan.Application)
			warningsStream <- Warnings(routeWarnings)
			if routeErr != nil {
				errorStream <- routeErr
				return
			}
			eventStream <- CreatedRoutes
		}

		err = actor.ScaleProcess(plan, warningsStream, eventStream)
		if err != nil {
			errorStream <- err
			return
		}

		err = actor.UpdateProcess(plan, warningsStream, eventStream)
		if err != nil {
			errorStream <- err
			return
		}

		pkg, err := actor.CreatePackage(plan, progressBar, warningsStream, eventStream)
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

		if plan.NoStart == true {
			if plan.Application.State == constant.ApplicationStarted {
				eventStream <- StoppingApplication
				warnings, err = actor.V7Actor.StopApplication(plan.Application.GUID)
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

		droplet, warnings, err := actor.V7Actor.PollBuild(build.GUID, plan.Application.Name)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- StagingComplete
		eventStream <- SettingDroplet

		warnings, err = actor.V7Actor.SetApplicationDroplet(plan.Application.GUID, droplet.GUID)
		warningsStream <- Warnings(warnings)
		if err != nil {
			errorStream <- err
			return
		}

		eventStream <- SetDropletComplete

		log.Debug("completed apply")
		eventStream <- Complete
	}()
	return planStream, eventStream, warningsStream, errorStream
}

func (actor Actor) CreateAndUploadApplicationBits(plan PushPlan, progressBar ProgressBar, warningsStream chan Warnings, eventStream chan Event) (v7action.Package, error) {
	log.WithField("Path", plan.BitsPath).Info("creating archive")

	eventStream <- CreatingArchive
	archivePath, err := actor.GetArchivePath(plan)
	if err != nil {
		return v7action.Package{}, err
	}
	defer os.RemoveAll(archivePath)

	eventStream <- CreatingPackage
	log.WithField("GUID", plan.Application.GUID).Info("creating package")
	pkg, warnings, err := actor.V7Actor.CreateBitsPackageByApplication(plan.Application.GUID)
	warningsStream <- Warnings(warnings)
	if err != nil {
		return v7action.Package{}, err
	}

	// Uploading package/app bits
	for count := 0; count < PushRetries; count++ {
		eventStream <- ReadingArchive
		log.WithField("GUID", plan.Application.GUID).Info("reading archive")
		file, size, readErr := actor.SharedActor.ReadArchive(archivePath)
		if readErr != nil {
			return v7action.Package{}, readErr
		}
		defer file.Close()

		eventStream <- UploadingApplicationWithArchive
		progressReader := progressBar.NewProgressBarWrapper(file, size)
		pkg, warnings, err = actor.V7Actor.UploadBitsPackage(pkg, plan.MatchedResources, progressReader, size)
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

func (actor Actor) updateApplication(plan PushPlan, warningsStream chan Warnings) (PushPlan, error) {
	if !plan.ApplicationNeedsUpdate {
		return plan, nil
	}

	log.WithField("Name", plan.Application.Name).Info("updating app")

	application, warnings, err := actor.V7Actor.UpdateApplication(plan.Application)
	plan.Application = application
	warningsStream <- Warnings(warnings)
	if err != nil {
		return plan, err
	}

	return plan, nil
}

func (actor Actor) CreatePackage(plan PushPlan, progressBar ProgressBar, warningsStream chan Warnings, eventStream chan Event) (v7action.Package, error) {
	if plan.DockerImageCredentialsNeedsUpdate {
		eventStream <- SetDockerImage
		pkg, warnings, err := actor.V7Actor.CreateDockerPackageByApplication(plan.Application.GUID, plan.DockerImageCredentials)
		warningsStream <- Warnings(warnings)
		if err != nil {
			return v7action.Package{}, err
		}
		eventStream <- SetDockerImageComplete
		return pkg, nil
	}

	return actor.CreateAndUploadApplicationBits(plan, progressBar, warningsStream, eventStream)
}

func (actor Actor) GetArchivePath(plan PushPlan) (string, error) {
	if plan.Archive {
		return actor.SharedActor.ZipArchiveResources(plan.BitsPath, plan.AllResources)
	}
	return actor.SharedActor.ZipDirectoryResources(plan.BitsPath, plan.AllResources)
}

func (actor Actor) ScaleProcess(plan PushPlan, warningsStream chan Warnings, eventStream chan Event) error {
	if shouldScaleProcess(plan) {
		log.Info("Scaling Web Process")
		eventStream <- ScaleWebProcess

		process := v7action.Process{
			Type:       constant.ProcessTypeWeb,
			MemoryInMB: plan.Overrides.Memory,
			DiskInMB:   plan.Overrides.Disk,
			Instances:  plan.Overrides.Instances,
		}
		scaleWarnings, err := actor.V7Actor.ScaleProcessByApplication(plan.Application.GUID, process)
		warningsStream <- Warnings(scaleWarnings)
		if err != nil {
			return err
		}
		eventStream <- ScaleWebProcessComplete
	}

	return nil
}

func shouldScaleProcess(plan PushPlan) bool {
	return plan.Overrides.Memory.IsSet || plan.Overrides.Instances.IsSet || plan.Overrides.Disk.IsSet
}

func (actor Actor) UpdateProcess(plan PushPlan, warningsStream chan Warnings, eventStream chan Event) error {
	if plan.Overrides.StartCommand.IsSet || plan.Overrides.HealthCheckType != "" || plan.Overrides.HealthCheckTimeout != 0 {
		log.Info("Setting Web Process's Configuration")
		eventStream <- SetProcessConfiguration

		var process v7action.Process
		if plan.Overrides.StartCommand.IsSet {
			process.Command = plan.Overrides.StartCommand
		}
		if plan.Overrides.HealthCheckType != "" {
			process.HealthCheckType = plan.Overrides.HealthCheckType
			process.HealthCheckEndpoint = plan.Overrides.HealthCheckEndpoint
		}
		if plan.Overrides.HealthCheckTimeout != 0 {
			process.HealthCheckTimeout = plan.Overrides.HealthCheckTimeout
		}

		log.WithField("Process", process).Debug("Update process")
		warnings, err := actor.V7Actor.UpdateProcessByTypeAndApplication(constant.ProcessTypeWeb, plan.Application.GUID, process)
		warningsStream <- Warnings(warnings)
		if err != nil {
			return err
		}
		eventStream <- SetProcessConfigurationComplete
	}

	return nil
}
