package v3action

import (
	"os"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

const (
	DefaultFolderPermissions      = 0755
	DefaultArchiveFilePermissions = 0744
)

type Package ccv3.Package

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
	inputPackage := ccv3.Package{
		Type: constant.PackageTypeDocker,
		Relationships: ccv3.Relationships{
			ccv3.ApplicationRelationship: ccv3.Relationship{GUID: app.GUID},
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

	var resources []sharedaction.Resource
	if info.IsDir() {
		resources, err = actor.SharedActor.GatherDirectoryResources(bitsPath)
	} else {
		resources, err = actor.SharedActor.GatherArchiveResources(bitsPath)
	}
	if err != nil {
		return Package{}, allWarnings, err
	}

	// potentially match resources here in the future

	var archivePath string
	if info.IsDir() {
		archivePath, err = actor.SharedActor.ZipDirectoryResources(bitsPath, resources)
	} else {
		archivePath, err = actor.SharedActor.ZipArchiveResources(bitsPath, resources)
	}
	if err != nil {
		return Package{}, allWarnings, err
	}

	inputPackage := ccv3.Package{
		Type: constant.PackageTypeBits,
		Relationships: ccv3.Relationships{
			ccv3.ApplicationRelationship: ccv3.Relationship{GUID: app.GUID},
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

	return Package(pkg), allWarnings, err
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
