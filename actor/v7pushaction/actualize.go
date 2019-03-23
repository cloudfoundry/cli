package v7pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"

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

		if plan.NoStart {
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
	var v7warnings v7action.Warnings

	eventStream <- ResourceMatching
	matchedResources, unmatchedResources, warnings, err := actor.MatchResources(plan.AllResources)
	warningsStream <- warnings
	if err != nil {
		return v7action.Package{}, err
	}

	eventStream <- CreatingPackage
	log.WithField("GUID", plan.Application.GUID).Info("creating package")
	pkg, v7warnings, err := actor.V7Actor.CreateBitsPackageByApplication(plan.Application.GUID)
	warningsStream <- Warnings(v7warnings)
	if err != nil {
		return v7action.Package{}, err
	}

	if len(unmatchedResources) > 0 {
		eventStream <- CreatingArchive
		archivePath, err := actor.GetArchivePath(plan, unmatchedResources)
		if err != nil {
			return v7action.Package{}, err
		}
		defer os.RemoveAll(archivePath)

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
			pkg, v7warnings, err = actor.V7Actor.UploadBitsPackage(pkg, matchedResources, progressReader, size)
			warningsStream <- Warnings(v7warnings)

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
	} else {
		eventStream <- UploadingApplication
		pkg, v7warnings, err = actor.V7Actor.UploadBitsPackage(pkg, matchedResources, nil, 0)

		warningsStream <- Warnings(v7warnings)
		if err != nil {
			return v7action.Package{}, err
		}
	}
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

func (actor Actor) GetArchivePath(plan PushPlan, unmatchedResources []sharedaction.V3Resource) (string, error) {
	// translate between v3 and v2 resources
	var v2Resources []sharedaction.Resource
	for _, resource := range unmatchedResources {
		v2Resources = append(v2Resources, resource.ToV2Resource())
	}

	if plan.Archive {
		return actor.SharedActor.ZipArchiveResources(plan.BitsPath, v2Resources)
	}
	return actor.SharedActor.ZipDirectoryResources(plan.BitsPath, v2Resources)
}

func (actor Actor) ScaleProcess(plan PushPlan, warningsStream chan Warnings, eventStream chan Event) error {
	if plan.ScaleWebProcessNeedsUpdate {
		log.Info("Scaling Web Process")
		eventStream <- ScaleWebProcess

		scaleWarnings, err := actor.V7Actor.ScaleProcessByApplication(plan.Application.GUID, plan.ScaleWebProcess)
		warningsStream <- Warnings(scaleWarnings)
		if err != nil {
			return err
		}
		eventStream <- ScaleWebProcessComplete
	}

	return nil
}

func (actor Actor) UpdateProcess(plan PushPlan, warningsStream chan Warnings, eventStream chan Event) error {
	if plan.UpdateWebProcessNeedsUpdate {
		log.Info("Setting Web Process's Configuration")
		eventStream <- SetProcessConfiguration

		log.WithField("Process", plan.UpdateWebProcess).Debug("Update process")
		warnings, err := actor.V7Actor.UpdateProcessByTypeAndApplication(constant.ProcessTypeWeb, plan.Application.GUID, plan.UpdateWebProcess)
		warningsStream <- Warnings(warnings)
		if err != nil {
			return err
		}
		eventStream <- SetProcessConfigurationComplete
	}

	return nil
}
