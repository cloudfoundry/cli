package v3action

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Build ccv3.Build

type StagingTimeoutError struct {
	AppName string
	Timeout time.Duration
}

func (StagingTimeoutError) Error() string {
	return "Timed out waiting for package to stage"
}

func (actor Actor) StagePackage(packageGUID string, appName string) (<-chan Build, <-chan Warnings, <-chan error) {
	buildStream := make(chan Build)
	warningsStream := make(chan Warnings)
	errorStream := make(chan error)

	go func() {
		defer close(buildStream)
		defer close(warningsStream)
		defer close(errorStream)

		build := ccv3.Build{Package: ccv3.Package{GUID: packageGUID}}
		build, allWarnings, err := actor.CloudControllerClient.CreateBuild(build)
		warningsStream <- Warnings(allWarnings)

		if err != nil {
			errorStream <- err
			return
		}

		timeout := time.Now().Add(actor.Config.StagingTimeout())

		for time.Now().Before(timeout) {
			var warnings ccv3.Warnings
			build, warnings, err = actor.CloudControllerClient.GetBuild(build.GUID)
			warningsStream <- Warnings(warnings)
			if err != nil {
				errorStream <- err
				return
			}

			switch build.State {
			case ccv3.BuildStateFailed:
				errorStream <- errors.New(build.Error)
				return
			case ccv3.BuildStateStaging:
				time.Sleep(actor.Config.PollingInterval())
			default:
				buildStream <- Build(build)
				return
			}
		}

		errorStream <- StagingTimeoutError{AppName: appName, Timeout: actor.Config.StagingTimeout()}
	}()

	return buildStream, warningsStream, errorStream
}
