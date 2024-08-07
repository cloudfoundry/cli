package v7pushaction

import (
	"io"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/resources"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . V7Actor

type V7Actor interface {
	CreateApplicationDroplet(appGUID string) (resources.Droplet, v7action.Warnings, error)
	CreateApplicationInSpace(app resources.Application, spaceGUID string) (resources.Application, v7action.Warnings, error)
	CreateBitsPackageByApplication(appGUID string) (resources.Package, v7action.Warnings, error)
	CreateDeployment(dep resources.Deployment) (string, v7action.Warnings, error)
	CreateDockerPackageByApplication(appGUID string, dockerImageCredentials v7action.DockerImageCredentials) (resources.Package, v7action.Warnings, error)
	CreateRoute(spaceGUID, domainName, hostname, path string, port int) (resources.Route, v7action.Warnings, error)
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (resources.Application, v7action.Warnings, error)
	GetApplicationDroplets(appName string, spaceGUID string) ([]resources.Droplet, v7action.Warnings, error)
	GetApplicationRoutes(appGUID string) ([]resources.Route, v7action.Warnings, error)
	GetApplicationsByNamesAndSpace(appNames []string, spaceGUID string) ([]resources.Application, v7action.Warnings, error)
	GetDefaultDomain(orgGUID string) (resources.Domain, v7action.Warnings, error)
	GetDomain(domainGUID string) (resources.Domain, v7action.Warnings, error)
	GetRouteByAttributes(domain resources.Domain, hostname, path string, port int) (resources.Route, v7action.Warnings, error)
	GetRouteDestinationByAppGUID(route resources.Route, appGUID string) (resources.RouteDestination, error)
	MapRoute(routeGUID string, appGUID string, destinationProtocol string) (v7action.Warnings, error)
	PollBuild(buildGUID string, appName string) (resources.Droplet, v7action.Warnings, error)
	PollPackage(pkg resources.Package) (resources.Package, v7action.Warnings, error)
	PollStart(app resources.Application, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	PollStartForDeployment(app resources.Application, deploymentGUID string, noWait bool, handleProcessStats func(string)) (v7action.Warnings, error)
	ResourceMatch(resources []sharedaction.V3Resource) ([]sharedaction.V3Resource, v7action.Warnings, error)
	RestartApplication(appGUID string, noWait bool) (v7action.Warnings, error)
	ScaleProcessByApplication(appGUID string, process resources.Process) (v7action.Warnings, error)
	SetApplicationDroplet(appGUID string, dropletGUID string) (v7action.Warnings, error)
	SetApplicationManifest(appGUID string, rawManifest []byte) (v7action.Warnings, error)
	SetSpaceManifest(spaceGUID string, rawManifest []byte) (v7action.Warnings, error)
	StageApplicationPackage(pkgGUID string) (resources.Build, v7action.Warnings, error)
	StopApplication(appGUID string) (v7action.Warnings, error)
	UnmapRoute(routeGUID string, destinationGUID string) (v7action.Warnings, error)
	UpdateApplication(app resources.Application) (resources.Application, v7action.Warnings, error)
	UpdateProcessByTypeAndApplication(processType string, appGUID string, updatedProcess resources.Process) (v7action.Warnings, error)
	UploadBitsPackage(pkg resources.Package, matchedResources []sharedaction.V3Resource, newResources io.Reader, newResourcesLength int64) (resources.Package, v7action.Warnings, error)
	UploadDroplet(dropletGUID string, dropletPath string, progressReader io.Reader, fileSize int64) (v7action.Warnings, error)
}
