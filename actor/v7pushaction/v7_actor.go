package v7pushaction

import (
	"io"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
)

//go:generate counterfeiter . V7Actor

type V7Actor interface {
	CreateApplicationDroplet(appGUID string) (v7action.Droplet, v7action.Warnings, error)
	CreateApplicationInSpace(app v7action.Application, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	CreateBitsPackageByApplication(appGUID string) (v7action.Package, v7action.Warnings, error)
	CreateDeployment(appGUID string, dropletGUID string) (string, v7action.Warnings, error)
	CreateDockerPackageByApplication(appGUID string, dockerImageCredentials v7action.DockerImageCredentials) (v7action.Package, v7action.Warnings, error)
	CreateRoute(spaceGUID, domainName, hostname, path string) (resources.Route, v7action.Warnings, error)
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v7action.Application, v7action.Warnings, error)
	GetApplicationDroplets(appName string, spaceGUID string) ([]v7action.Droplet, v7action.Warnings, error)
	GetApplicationRoutes(appGUID string) ([]resources.Route, v7action.Warnings, error)
	GetApplicationsByNamesAndSpace(appNames []string, spaceGUID string) ([]v7action.Application, v7action.Warnings, error)
	GetDefaultDomain(orgGUID string) (v7action.Domain, v7action.Warnings, error)
	GetDomain(domainGUID string) (v7action.Domain, v7action.Warnings, error)
	GetRouteByAttributes(domainName, domainGUID, hostname, path string) (resources.Route, v7action.Warnings, error)
	GetRouteDestinationByAppGUID(routeGUID string, appGUID string) (resources.RouteDestination, v7action.Warnings, error)
	MapRoute(routeGUID string, appGUID string) (v7action.Warnings, error)
	PollBuild(buildGUID string, appName string) (v7action.Droplet, v7action.Warnings, error)
	PollPackage(pkg v7action.Package) (v7action.Package, v7action.Warnings, error)
	PollStart(appGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	PollStartForRolling(appGUID string, deploymentGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	ResourceMatch(resources []sharedaction.V3Resource) ([]sharedaction.V3Resource, v7action.Warnings, error)
	RestartApplication(appGUID string, noWait bool) (v7action.Warnings, error)
	ScaleProcessByApplication(appGUID string, process v7action.Process) (v7action.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (v7action.Warnings, error)
	SetApplicationManifest(appGUID string, rawManifest []byte) (v7action.Warnings, error)
	SetSpaceManifest(spaceGUID string, rawManifest []byte) (v7action.Warnings, error)
	StageApplicationPackage(pkgGUID string) (v7action.Build, v7action.Warnings, error)
	StopApplication(appGUID string) (v7action.Warnings, error)
	UnmapRoute(routeGUID string, destinationGUID string) (v7action.Warnings, error)
	UpdateApplication(app v7action.Application) (v7action.Application, v7action.Warnings, error)
	UpdateProcessByTypeAndApplication(processType string, appGUID string, updatedProcess v7action.Process) (v7action.Warnings, error)
	UploadBitsPackage(pkg v7action.Package, matchedResources []sharedaction.V3Resource, newResources io.Reader, newResourcesLength int64) (v7action.Package, v7action.Warnings, error)
	UploadDroplet(dropletGUID string, dropletPath string, progressReader io.Reader, fileSize int64) (v7action.Warnings, error)
}
