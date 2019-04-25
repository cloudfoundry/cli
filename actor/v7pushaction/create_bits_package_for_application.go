package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	log "github.com/sirupsen/logrus"
	"os"
)

const PushRetries = 3

func (actor Actor) CreateBitsPackageForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	if !pushPlan.DockerImageCredentialsNeedsUpdate {
		pkg, warnings, err := actor.CreateAndUploadApplicationBits(pushPlan, progressBar, eventStream)
		if err != nil {
			return pushPlan, warnings, err
		}

		polledPackage, pollWarnings, err := actor.V7Actor.PollPackage(pkg)

		pushPlan.PackageGUID = polledPackage.GUID

		return pushPlan, append(warnings, pollWarnings...), err
	}

	return pushPlan, nil, nil
}

func (actor Actor) CreateAndUploadApplicationBits(plan PushPlan, progressBar ProgressBar, eventStream chan<- Event) (v7action.Package, Warnings, error) {
	log.WithField("Path", plan.BitsPath).Info("creating archive")
	var allWarnings Warnings

	eventStream <- ResourceMatching
	matchedResources, unmatchedResources, warnings, err := actor.MatchResources(plan.AllResources)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return v7action.Package{}, allWarnings, err
	}

	eventStream <- CreatingPackage
	log.WithField("GUID", plan.Application.GUID).Info("creating package")
	pkg, createPackageWarnings, err := actor.V7Actor.CreateBitsPackageByApplication(plan.Application.GUID)
	allWarnings = append(allWarnings, createPackageWarnings...)
	if err != nil {
		return v7action.Package{}, allWarnings, err
	}

	if len(unmatchedResources) > 0 {
		eventStream <- CreatingArchive
		archivePath, archiveErr := actor.CreateAndReturnArchivePath(plan, unmatchedResources)
		if archiveErr != nil {
			return v7action.Package{}, allWarnings, archiveErr
		}
		defer os.RemoveAll(archivePath)

		// Uploading package/app bits
		for count := 0; count < PushRetries; count++ {
			eventStream <- ReadingArchive
			log.WithField("GUID", plan.Application.GUID).Info("reading archive")
			file, size, readErr := actor.SharedActor.ReadArchive(archivePath)
			if readErr != nil {
				return v7action.Package{}, allWarnings, readErr
			}
			defer file.Close()

			eventStream <- UploadingApplicationWithArchive
			progressReader := progressBar.NewProgressBarWrapper(file, size)
			var uploadWarnings v7action.Warnings
			pkg, uploadWarnings, err = actor.V7Actor.UploadBitsPackage(pkg, matchedResources, progressReader, size)
			allWarnings = append(allWarnings, uploadWarnings...)

			if _, ok := err.(ccerror.PipeSeekError); ok {
				eventStream <- RetryUpload
				continue
			}
			break
		}

		if err != nil {
			if e, ok := err.(ccerror.PipeSeekError); ok {
				return v7action.Package{}, allWarnings, actionerror.UploadFailedError{Err: e.Err}
			}
			return v7action.Package{}, allWarnings, err
		}

		eventStream <- UploadWithArchiveComplete
	} else {
		eventStream <- UploadingApplication
		var uploadWarnings v7action.Warnings
		pkg, uploadWarnings, err = actor.V7Actor.UploadBitsPackage(pkg, matchedResources, nil, 0)
		allWarnings = append(allWarnings, uploadWarnings...)
		if err != nil {
			return v7action.Package{}, allWarnings, err
		}
	}
	return pkg, allWarnings, nil
}

func (actor Actor) CreateAndReturnArchivePath(plan PushPlan, unmatchedResources []sharedaction.V3Resource) (string, error) {
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
