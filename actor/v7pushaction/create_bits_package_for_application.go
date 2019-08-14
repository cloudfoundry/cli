package v7pushaction

import (
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	log "github.com/sirupsen/logrus"
)

const PushRetries = 3

func (actor Actor) CreateBitsPackageForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	pkg, warnings, err := actor.CreateAndUploadApplicationBits(pushPlan, eventStream, progressBar)
	if err != nil {
		return pushPlan, warnings, err
	}

	polledPackage, pollWarnings, err := actor.V7Actor.PollPackage(pkg)

	pushPlan.PackageGUID = polledPackage.GUID

	return pushPlan, append(warnings, pollWarnings...), err
}

func (actor Actor) CreateAndUploadApplicationBits(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (v7action.Package, Warnings, error) {
	log.WithField("Path", pushPlan.BitsPath).Info("creating archive")

	var (
		allWarnings        Warnings
		matchedResources   []sharedaction.V3Resource
		unmatchedResources []sharedaction.V3Resource
	)

	// check if all source files are empty
	shouldResourceMatch := false
	for _, resource := range pushPlan.AllResources {
		if resource.SizeInBytes != 0 {
			shouldResourceMatch = true
		}
	}

	if shouldResourceMatch {
		eventStream <- &PushEvent{Plan: pushPlan, Event: ResourceMatching}
		var warnings Warnings
		var err error

		matchedResources, unmatchedResources, warnings, err = actor.MatchResources(pushPlan.AllResources)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return v7action.Package{}, allWarnings, err
		}
	} else {
		matchedResources = []sharedaction.V3Resource{}
		unmatchedResources = pushPlan.AllResources
	}

	eventStream <- &PushEvent{Plan: pushPlan, Event: CreatingPackage}
	log.WithField("GUID", pushPlan.Application.GUID).Info("creating package")
	pkg, createPackageWarnings, err := actor.V7Actor.CreateBitsPackageByApplication(pushPlan.Application.GUID)
	allWarnings = append(allWarnings, createPackageWarnings...)
	if err != nil {
		return v7action.Package{}, allWarnings, err
	}

	if len(unmatchedResources) > 0 {
		eventStream <- &PushEvent{Plan: pushPlan, Event: CreatingArchive}
		archivePath, archiveErr := actor.CreateAndReturnArchivePath(pushPlan, unmatchedResources)
		if archiveErr != nil {
			return v7action.Package{}, allWarnings, archiveErr
		}
		defer os.RemoveAll(archivePath)

		// Uploading package/app bits
		for count := 0; count < PushRetries; count++ {
			eventStream <- &PushEvent{Plan: pushPlan, Event: ReadingArchive}
			log.WithField("GUID", pushPlan.Application.GUID).Info("reading archive")
			file, size, readErr := actor.SharedActor.ReadArchive(archivePath)
			if readErr != nil {
				return v7action.Package{}, allWarnings, readErr
			}
			defer file.Close()

			eventStream <- &PushEvent{Plan: pushPlan, Event: UploadingApplicationWithArchive}
			progressReader := progressBar.NewProgressBarWrapper(file, size)
			var uploadWarnings v7action.Warnings
			pkg, uploadWarnings, err = actor.V7Actor.UploadBitsPackage(pkg, matchedResources, progressReader, size)
			allWarnings = append(allWarnings, uploadWarnings...)

			if _, ok := err.(ccerror.PipeSeekError); ok {
				eventStream <- &PushEvent{Plan: pushPlan, Event: RetryUpload}
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

		eventStream <- &PushEvent{Plan: pushPlan, Event: UploadWithArchiveComplete}
	} else {
		eventStream <- &PushEvent{Plan: pushPlan, Event: UploadingApplication}
		var uploadWarnings v7action.Warnings
		pkg, uploadWarnings, err = actor.V7Actor.UploadBitsPackage(pkg, matchedResources, nil, 0)
		allWarnings = append(allWarnings, uploadWarnings...)
		if err != nil {
			return v7action.Package{}, allWarnings, err
		}
	}

	return pkg, allWarnings, nil
}

func (actor Actor) CreateAndReturnArchivePath(pushPlan PushPlan, unmatchedResources []sharedaction.V3Resource) (string, error) {
	// translate between v3 and v2 resources
	var v2Resources []sharedaction.Resource
	for _, resource := range unmatchedResources {
		v2Resources = append(v2Resources, resource.ToV2Resource())
	}

	if pushPlan.Archive {
		return actor.SharedActor.ZipArchiveResources(pushPlan.BitsPath, v2Resources)
	}
	return actor.SharedActor.ZipDirectoryResources(pushPlan.BitsPath, v2Resources)
}
