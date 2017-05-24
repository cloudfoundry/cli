package v3action

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Build ccv3.Build

func (actor Actor) StagePackage(packageGUID string) (Build, Warnings, error) {
	build := ccv3.Build{Package: ccv3.Package{GUID: packageGUID}}
	build, allWarnings, err := actor.CloudControllerClient.CreateBuild(build)

	if err != nil {
		return Build{}, Warnings(allWarnings), err
	}

	for build.State == ccv3.BuildStateStaging {
		time.Sleep(actor.Config.PollingInterval())

		var warnings ccv3.Warnings
		build, warnings, err = actor.CloudControllerClient.GetBuild(build.GUID)
		allWarnings = append(allWarnings, warnings...)
	}
	return Build(build), Warnings(allWarnings), err
}
