package v3action

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Build ccv3.Build

func (actor Actor) StagePackage(packageGUID string) (<-chan Build, <-chan Warnings, <-chan error) {
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

		for build.State == ccv3.BuildStateStaging {
			time.Sleep(actor.Config.PollingInterval())

			var warnings ccv3.Warnings
			build, warnings, err = actor.CloudControllerClient.GetBuild(build.GUID)
			warningsStream <- Warnings(warnings)
			if err != nil {
				errorStream <- err
				return
			}
		}
		buildStream <- Build(build)
	}()

	return buildStream, warningsStream, errorStream
}
