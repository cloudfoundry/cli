package v3action

import (
	"io"
	"os"
	"time"

	"code.cloudfoundry.org/cli/v7/actor/actionerror"
	"code.cloudfoundry.org/cli/v7/actor/sharedaction"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v7/resources"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultFolderPermissions      = 0755
	DefaultArchiveFilePermissions = 0744
)

type Package resources.Package

type DockerImageCredentials struct {
	Path     string
	Username string
	Password string
}

func (actor Actor) CreateDockerPackageByApplicationNameAndSpace(appName string, spaceGUID string, dockerImageCredentials DockerImageCredentials) (Package, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return Package{}, allWarnings, err
	}
	inputPackage := resources.Package{
		Type: constant.PackageTypeDocker,
		Relationships: resources.Relationships{
			constant.RelationshipTypeApplication: resources.Relationship{GUID: app.GUID},
		},
		DockerImage:    dockerImageCredentials.Path,
		DockerUsername: dockerImageCredentials.Username,
		DockerPassword: dockerImageCredentials.Password,
	}
	pkg, warnings, err := actor.CloudControllerClient.CreatePackage(inputPackage)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Package{}, allWarnings, err
	}
	return Package(pkg), allWarnings, err
}

func (actor Actor) CreateAndUploadBitsPackageByApplicationNameAndSpace(appName string, spaceGUID string, bitsPath string) (Package, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return Package{}, allWarnings, err
	}

	if bitsPath == "" {
		bitsPath, err = os.Getwd()
		if err != nil {
			return Package{}, allWarnings, err
		}
	}

	info, err := os.Stat(bitsPath)
	if err != nil {
		return Package{}, allWarnings, err
	}

	var fileResources []sharedaction.Resource
	if info.IsDir() {
		fileResources, err = actor.SharedActor.GatherDirectoryResources(bitsPath)
	} else {
		fileResources, err = actor.SharedActor.GatherArchiveResources(bitsPath)
	}
	if err != nil {
		return Package{}, allWarnings, err
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
		return Package{}, allWarnings, err
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
		return Package{}, allWarnings, err
	}

	_, warnings, err = actor.CloudControllerClient.UploadPackage(pkg, archivePath)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Package{}, allWarnings, err
	}

	for pkg.State != constant.PackageReady &&
		pkg.State != constant.PackageFailed &&
		pkg.State != constant.PackageExpired {
		time.Sleep(actor.Config.PollingInterval())
		pkg, warnings, err = actor.CloudControllerClient.GetPackage(pkg.GUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return Package{}, allWarnings, err
		}
	}

	if pkg.State == constant.PackageFailed {
		return Package{}, allWarnings, actionerror.PackageProcessingFailedError{}
	} else if pkg.State == constant.PackageExpired {
		return Package{}, allWarnings, actionerror.PackageProcessingExpiredError{}
	}

	updatedPackage, updatedWarnings, err := actor.PollPackage(Package(pkg))
	return updatedPackage, append(allWarnings, updatedWarnings...), err
}

// GetApplicationPackages returns a list of package of an app.
func (actor *Actor) GetApplicationPackages(appName string, spaceGUID string) ([]Package, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	ccv3Packages, warnings, err := actor.CloudControllerClient.GetPackages(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{app.GUID}},
	)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var packages []Package
	for _, ccv3Package := range ccv3Packages {
		packages = append(packages, Package(ccv3Package))
	}

	return packages, allWarnings, nil
}

func (actor Actor) CreateBitsPackageByApplication(appGUID string) (Package, Warnings, error) {
	inputPackage := resources.Package{
		Type: constant.PackageTypeBits,
		Relationships: resources.Relationships{
			constant.RelationshipTypeApplication: resources.Relationship{GUID: appGUID},
		},
	}

	pkg, warnings, err := actor.CloudControllerClient.CreatePackage(inputPackage)
	if err != nil {
		return Package{}, Warnings(warnings), err
	}

	return Package(pkg), Warnings(warnings), err
}

func (actor Actor) UploadBitsPackage(pkg Package, existingResources []sharedaction.Resource, newResources io.Reader, newResourcesLength int64) (Package, Warnings, error) {
	apiResources := make([]ccv3.Resource, 0, len(existingResources)) // Explicitly done to prevent nils

	for _, resource := range existingResources {
		apiResources = append(apiResources, ccv3.Resource(resource.ToV3Resource()))
	}

	appPkg, warnings, err := actor.CloudControllerClient.UploadBitsPackage(resources.Package(pkg), apiResources, newResources, newResourcesLength)
	return Package(appPkg), Warnings(warnings), err
}

// PollPackage returns a package of an app.
func (actor Actor) PollPackage(pkg Package) (Package, Warnings, error) {
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
			return Package{}, allWarnings, err
		}

		pkg = Package(ccPkg)
	}

	if pkg.State == constant.PackageFailed {
		return Package{}, allWarnings, actionerror.PackageProcessingFailedError{}
	} else if pkg.State == constant.PackageExpired {
		return Package{}, allWarnings, actionerror.PackageProcessingExpiredError{}
	}

	return pkg, allWarnings, nil
}
