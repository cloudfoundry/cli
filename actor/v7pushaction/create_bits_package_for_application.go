package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/resources"
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

func (actor Actor) CreateAndUploadApplicationBits(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (resources.Package, Warnings, error) {
	log.WithField("Path", pushPlan.BitsPath).Info("creating archive")

	var (
		allWarnings Warnings
	)

	eventStream <- &PushEvent{Plan: pushPlan, Event: CreatingPackage}
	log.WithField("GUID", pushPlan.Application.GUID).Info("creating package")
	pkg, createPackageWarnings, err := actor.V7Actor.CreateAndUploadBitsPackageByApplicationNameAndSpace(pushPlan.Application.Name, pushPlan.SpaceGUID, pushPlan.BitsPath)
	allWarnings = append(allWarnings, createPackageWarnings...)
	if err != nil {
		return resources.Package{}, allWarnings, err
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
