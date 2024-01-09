package v7action

import (
	"io"
	"os"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultFolderPermissions      = 0755
	DefaultArchiveFilePermissions = 0744
)

type DockerImageCredentials struct {
	Path     string
	Username string
	Password string
}

func (actor Actor) CreateDockerPackageByApplication(appGUID string, dockerImageCredentials DockerImageCredentials) (resources.Package, Warnings, error) {
	inputPackage := resources.Package{
		Type: constant.PackageTypeDocker,
		Relationships: resources.Relationships{
			constant.RelationshipTypeApplication: resources.Relationship{GUID: appGUID},
		},
		DockerImage:    dockerImageCredentials.Path,
		DockerUsername: dockerImageCredentials.Username,
		DockerPassword: dockerImageCredentials.Password,
	}
	pkg, warnings, err := actor.CloudControllerClient.CreatePackage(inputPackage)
	return resources.Package(pkg), Warnings(warnings), err
}

func (actor Actor) CreateDockerPackageByApplicationNameAndSpace(appName string, spaceGUID string, dockerImageCredentials DockerImageCredentials) (resources.Package, Warnings, error) {
	app, getWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return resources.Package{}, getWarnings, err
	}
	pkg, warnings, err := actor.CreateDockerPackageByApplication(app.GUID, dockerImageCredentials)
	return pkg, append(getWarnings, warnings...), err
}

func (actor Actor) CreateAndUploadBitsPackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string) (resources.Package, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return resources.Package{}, allWarnings, err
	}

	if bitsPath == "" {
		bitsPath, err = os.Getwd()
		if err != nil {
			return resources.Package{}, allWarnings, err
		}
	}

	info, err := os.Stat(bitsPath)
	if err != nil {
		return resources.Package{}, allWarnings, err
	}

	var fileResources []sharedaction.Resource
	if info.IsDir() {
		fileResources, err = actor.SharedActor.GatherDirectoryResources(bitsPath)
	} else {
		fileResources, err = actor.SharedActor.GatherArchiveResources(bitsPath)
	}
	if err != nil {
		return resources.Package{}, allWarnings, err
	}

	// potentially match resources here in the future

	var archivePath string
	if info.IsDir() {
		archivePath, err = actor.SharedActor.ZipDirectoryResources(bitsPath, fileResources)
	} else {
		archivePath, err = actor.SharedActor.ZipArchiveResources(bitsPath, fileResources)
	}
	if err != nil {
		os.RemoveAll(archivePath)
		return resources.Package{}, allWarnings, err
	}
	defer os.RemoveAll(archivePath)

	inputPackage := resources.Package{
		Type: constant.PackageTypeBits,
		Relationships: resources.Relationships{
			constant.RelationshipTypeApplication: resources.Relationship{GUID: app.GUID},
		},
	}

	pkg, warnings, err := actor.CloudControllerClient.CreatePackage(inputPackage)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Package{}, allWarnings, err
	}

	_, warnings, err = actor.CloudControllerClient.UploadPackage(pkg, archivePath)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Package{}, allWarnings, err
	}

	for pkg.State != constant.PackageReady &&
		pkg.State != constant.PackageFailed &&
		pkg.State != constant.PackageExpired {
		time.Sleep(actor.Config.PollingInterval())
		pkg, warnings, err = actor.CloudControllerClient.GetPackage(pkg.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return resources.Package{}, allWarnings, err
		}
	}

	if pkg.State == constant.PackageFailed {
		return resources.Package{}, allWarnings, actionerror.PackageProcessingFailedError{}
	} else if pkg.State == constant.PackageExpired {
		return resources.Package{}, allWarnings, actionerror.PackageProcessingExpiredError{}
	}

	updatedPackage, updatedWarnings, err := actor.PollPackage(resources.Package(pkg))
	return updatedPackage, append(allWarnings, updatedWarnings...), err
}

func (actor Actor) GetNewestReadyPackageForApplication(app resources.Application) (resources.Package, Warnings, error) {
	ccv3Packages, warnings, err := actor.CloudControllerClient.GetPackages(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{app.GUID}},
		ccv3.Query{Key: ccv3.StatesFilter, Values: []string{string(constant.PackageReady)}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)

	if err != nil {
		return resources.Package{}, Warnings(warnings), err
	}

	if len(ccv3Packages) == 0 {
		return resources.Package{}, Warnings(warnings), actionerror.NoEligiblePackagesError{AppName: app.Name}
	}

	return resources.Package(ccv3Packages[0]), Warnings(warnings), nil
}

// GetApplicationPackages returns a list of package of an app.
func (actor *Actor) GetApplicationPackages(appName string, spaceGUID string) ([]resources.Package, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	ccv3Packages, warnings, err := actor.CloudControllerClient.GetPackages(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{app.GUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
	)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var packages []resources.Package
	for _, ccv3Package := range ccv3Packages {
		packages = append(packages, resources.Package(ccv3Package))
	}

	return packages, allWarnings, nil
}

func (actor Actor) CreateBitsPackageByApplication(appGUID string) (resources.Package, Warnings, error) {
	inputPackage := resources.Package{
		Type: constant.PackageTypeBits,
		Relationships: resources.Relationships{
			constant.RelationshipTypeApplication: resources.Relationship{GUID: appGUID},
		},
	}

	pkg, warnings, err := actor.CloudControllerClient.CreatePackage(inputPackage)
	if err != nil {
		return resources.Package{}, Warnings(warnings), err
	}

	return resources.Package(pkg), Warnings(warnings), err
}

func (actor Actor) UploadBitsPackage(pkg resources.Package, matchedResources []sharedaction.V3Resource, newResources io.Reader, newResourcesLength int64) (resources.Package, Warnings, error) {
	apiResources := make([]ccv3.Resource, 0, len(matchedResources)) // Explicitly done to prevent nils

	for _, resource := range matchedResources {
		apiResources = append(apiResources, ccv3.Resource(resource))
	}

	appPkg, warnings, err := actor.CloudControllerClient.UploadBitsPackage(resources.Package(pkg), apiResources, newResources, newResourcesLength)
	return resources.Package(appPkg), Warnings(warnings), err
}

// PollPackage returns a package of an app.
func (actor Actor) PollPackage(pkg resources.Package) (resources.Package, Warnings, error) {
	var allWarnings Warnings

	for pkg.State != constant.PackageReady && pkg.State != constant.PackageFailed && pkg.State != constant.PackageExpired {
		time.Sleep(actor.Config.PollingInterval())
		ccPkg, warnings, err := actor.CloudControllerClient.GetPackage(pkg.GUID)
		log.WithFields(log.Fields{
			"package_guid": pkg.GUID,
			"state":        pkg.State,
		}).Debug("polling package state")

		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return resources.Package{}, allWarnings, err
		}

		pkg = resources.Package(ccPkg)
	}

	if pkg.State == constant.PackageFailed {
		return resources.Package{}, allWarnings, actionerror.PackageProcessingFailedError{}
	} else if pkg.State == constant.PackageExpired {
		return resources.Package{}, allWarnings, actionerror.PackageProcessingExpiredError{}
	}

	return pkg, allWarnings, nil
}

func (actor Actor) CopyPackage(sourceApp resources.Application, targetApp resources.Application) (resources.Package, Warnings, error) {
	var allWarnings Warnings
	sourcePkg, warnings, err := actor.GetNewestReadyPackageForApplication(sourceApp)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Package{}, allWarnings, err
	}
	targetPkg, ccv3Warnings, err := actor.CloudControllerClient.CopyPackage(sourcePkg.GUID, targetApp.GUID)
	allWarnings = append(allWarnings, ccv3Warnings...)
	if err != nil {
		return resources.Package{}, allWarnings, err
	}

	readyPackage, warnings, err := actor.PollPackage(resources.Package(targetPkg))
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return resources.Package{}, allWarnings, err
	}

	return readyPackage, allWarnings, nil
}
