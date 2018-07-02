package pushaction

import (
	"io"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
)

//go:generate counterfeiter . V3Actor

type V3Actor interface {
	CloudControllerAPIVersion() string
	CreateApplicationInSpace(app v3action.Application, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	CreateBitsPackageByApplication(appGUID string) (v3action.Package, v3action.Warnings, error)
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	PollPackage(pkg v3action.Package) (v3action.Package, v3action.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (v3action.Warnings, error)
	StageApplicationPackage(pkgGUID string) (v3action.Build, v3action.Warnings, error)
	PollBuild(buildGUID string, appName string) (v3action.Droplet, v3action.Warnings, error)
	UpdateApplication(v3action.Application) (v3action.Application, v3action.Warnings, error)
	UploadBitsPackage(v3action.Package, []sharedaction.Resource, io.Reader, int64) (v3action.Package, v3action.Warnings, error)
}
