package v2action

import (
	"io"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type Buildpack struct {
	GUID string
	// Name     string
	// Position int
	// Enabled  bool
}

//go:generate counterfeiter . ProgressBar

type ProgressBar interface {
	Initialize(string) io.Reader
	Terminate()
}

func (actor *Actor) CreateBuildpack(name string, position int, enabled bool) (Buildpack, Warnings, error) {

	buildpack := ccv2.Buildpack{
		Name:     name,
		Position: position,
		Enabled:  enabled,
	}

	ccBuildpack, warnings, err := actor.CloudControllerClient.CreateBuildpack(buildpack)
	if _, ok := err.(ccerror.BuildpackAlreadyExistsError); ok {
		return Buildpack{}, Warnings(warnings), actionerror.BuildpackAlreadyExistsError(name)
	}

	return Buildpack{GUID: ccBuildpack.GUID}, Warnings(warnings), err
}

func (actor *Actor) UploadBuildpack(GUID string, path string, pb ProgressBar) (Warnings, error) {
	progressBarReader := pb.Initialize(path)
	warnings, _ := actor.CloudControllerClient.UploadBuildpack(GUID, progressBarReader, 0)

	pb.Terminate()
	return Warnings(warnings), nil
}
