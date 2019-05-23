package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

const UploadRetries = 3

func (actor Actor) CreateDropletForApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	var allWarnings Warnings

	eventStream <- CreatingDroplet
	droplet, warnings, err := actor.V7Actor.CreateApplicationDroplet(pushPlan.Application.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return pushPlan, allWarnings, err
	}

	for count := 0; count < UploadRetries; count++ {
		eventStream <- ReadingArchive
		file, size, readErr := actor.SharedActor.ReadArchive(pushPlan.DropletPath)
		if readErr != nil {
			return pushPlan, allWarnings, readErr
		}
		defer file.Close()

		eventStream <- UploadingDroplet
		progressReader := progressBar.NewProgressBarWrapper(file, size)
		var uploadWarnings v7action.Warnings
		uploadWarnings, err = actor.V7Actor.UploadDroplet(droplet.GUID, pushPlan.DropletPath, progressReader, size)
		allWarnings = append(allWarnings, uploadWarnings...)

		if _, ok := err.(ccerror.PipeSeekError); ok {
			eventStream <- RetryUpload
			continue
		}

		break
	}

	if err != nil {
		if e, ok := err.(ccerror.PipeSeekError); ok {
			return pushPlan, allWarnings, actionerror.UploadFailedError{Err: e.Err}
		}

		return pushPlan, allWarnings, err
	}

	eventStream <- UploadDropletComplete
	pushPlan.DropletGUID = droplet.GUID

	return pushPlan, allWarnings, err
}
