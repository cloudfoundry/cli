package actors

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"code.cloudfoundry.org/cli/cf/api/applicationbits"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/appfiles"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/gofileutils/fileutils"
)

const windowsPathPrefix = `\\?\`

//go:generate counterfeiter . PushActor

type PushActor interface {
	UploadApp(appGUID string, zipFile *os.File, presentFiles []resources.AppFileResource) error
	ProcessPath(dirOrZipFile string, f func(string) error) error
	GatherFiles(localFiles []models.AppFileFields, appDir string, uploadDir string, useCache bool) ([]resources.AppFileResource, bool, error)
	ValidateAppParams(apps []models.AppParams) []error
	MapManifestRoute(routeName string, app models.Application, appParamsFromContext models.AppParams) error
}

type PushActorImpl struct {
	appBitsRepo applicationbits.Repository
	appfiles    appfiles.AppFiles
	zipper      appfiles.Zipper
	routeActor  RouteActor
}

func NewPushActor(appBitsRepo applicationbits.Repository, zipper appfiles.Zipper, appfiles appfiles.AppFiles, routeActor RouteActor) PushActor {
	return PushActorImpl{
		appBitsRepo: appBitsRepo,
		appfiles:    appfiles,
		zipper:      zipper,
		routeActor:  routeActor,
	}
}

// ProcessPath takes in a director of app files or a zip file which contains
// the app files. If given a zip file, it will extract the zip to a temporary
// location, call the provided callback with that location, and then clean up
// the location after the callback has been executed.
//
// This was done so that the caller of ProcessPath wouldn't need to know if it
// was a zip file or an app dir that it was given, and the caller would not be
// responsible for cleaning up the temporary directory ProcessPath creates when
// given a zip.
func (actor PushActorImpl) ProcessPath(dirOrZipFile string, f func(string) error) error {
	if !actor.zipper.IsZipFile(dirOrZipFile) {
		if filepath.IsAbs(dirOrZipFile) {
			appDir, err := filepath.EvalSymlinks(dirOrZipFile)
			if err != nil {
				return err
			}
			err = f(appDir)
			if err != nil {
				return err
			}
		} else {
			absPath, err := filepath.Abs(dirOrZipFile)
			if err != nil {
				return err
			}
			appDir, err := filepath.EvalSymlinks(absPath)
			if err != nil {
				return err
			}

			err = f(appDir)
			if err != nil {
				return err
			}
		}

		return nil
	}

	tempDir, err := ioutil.TempDir("", "unzipped-app")
	if err != nil {
		return err
	}

	err = actor.zipper.Unzip(dirOrZipFile, tempDir)
	if err != nil {
		return err
	}

	err = f(tempDir)
	if err != nil {
		return err
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		return err
	}

	return nil
}

func (actor PushActorImpl) GatherFiles(localFiles []models.AppFileFields, appDir string, uploadDir string, useCache bool) ([]resources.AppFileResource, bool, error) {
	appFileResource := []resources.AppFileResource{}
	for _, file := range localFiles {
		appFileResource = append(appFileResource, resources.AppFileResource{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		})
	}

	var err error
	// CC returns a list of files that it already has, so an empty list of
	// remoteFiles is equivalent to not using resource caching at all
	remoteFiles := []resources.AppFileResource{}
	if useCache {
		remoteFiles, err = actor.appBitsRepo.GetApplicationFiles(appFileResource)
		if err != nil {
			return []resources.AppFileResource{}, false, err
		}
	}

	filesToUpload := make([]models.AppFileFields, len(localFiles), len(localFiles))
	copy(filesToUpload, localFiles)

	for _, remoteFile := range remoteFiles {
		for i, fileToUpload := range filesToUpload {
			if remoteFile.Path == fileToUpload.Path {
				filesToUpload = append(filesToUpload[:i], filesToUpload[i+1:]...)
			}
		}
	}

	err = actor.appfiles.CopyFiles(filesToUpload, appDir, uploadDir)
	if err != nil {
		return []resources.AppFileResource{}, false, err
	}

	_, err = os.Stat(filepath.Join(appDir, ".cfignore"))
	if err == nil {
		err = fileutils.CopyPathToPath(filepath.Join(appDir, ".cfignore"), filepath.Join(uploadDir, ".cfignore"))
		if err != nil {
			return []resources.AppFileResource{}, false, err
		}
	}

	for i := range remoteFiles {
		fullPath, err := filepath.Abs(filepath.Join(appDir, remoteFiles[i].Path))
		if err != nil {
			return []resources.AppFileResource{}, false, err
		}

		if runtime.GOOS == "windows" {
			fullPath = windowsPathPrefix + fullPath
		}
		fileInfo, err := os.Lstat(fullPath)
		if err != nil {
			return []resources.AppFileResource{}, false, err
		}
		fileMode := fileInfo.Mode()

		if runtime.GOOS == "windows" {
			fileMode = fileMode | 0700
		}

		remoteFiles[i].Mode = fmt.Sprintf("%#o", fileMode)
	}

	return remoteFiles, len(filesToUpload) > 0, nil
}

func (actor PushActorImpl) UploadApp(appGUID string, zipFile *os.File, presentFiles []resources.AppFileResource) error {
	return actor.appBitsRepo.UploadBits(appGUID, zipFile, presentFiles)
}

func (actor PushActorImpl) ValidateAppParams(apps []models.AppParams) []error {
	errs := []error{}

	for _, app := range apps {
		appName := app.Name

		if app.HealthCheckType != nil && *app.HealthCheckType != "http" && app.HealthCheckHTTPEndpoint != nil {
			errs = append(errs, fmt.Errorf(T("Health check type must be 'http' to set a health check HTTP endpoint.")))
		}

		if app.Routes != nil {
			if app.Hosts != nil {
				errs = append(errs, fmt.Errorf(T("Application {{.AppName}} must not be configured with both 'routes' and 'host'/'hosts'", map[string]interface{}{"AppName": appName})))
			}

			if app.Domains != nil {
				errs = append(errs, fmt.Errorf(T("Application {{.AppName}} must not be configured with both 'routes' and 'domain'/'domains'", map[string]interface{}{"AppName": appName})))
			}

			if app.NoHostname != nil {
				errs = append(errs, fmt.Errorf(T("Application {{.AppName}} must not be configured with both 'routes' and 'no-hostname'", map[string]interface{}{"AppName": appName})))
			}
		}

		if app.BuildpackURL != nil && app.DockerImage != nil {
			errs = append(errs, fmt.Errorf(T("Application {{.AppName}} must not be configured with both 'buildpack' and 'docker'", map[string]interface{}{"AppName": appName})))
		}

		if app.Path != nil && app.DockerImage != nil {
			errs = append(errs, fmt.Errorf(T("Application {{.AppName}} must not be configured with both 'docker' and 'path'", map[string]interface{}{"AppName": appName})))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (actor PushActorImpl) MapManifestRoute(routeName string, app models.Application, appParamsFromContext models.AppParams) error {
	return actor.routeActor.FindAndBindRoute(routeName, app, appParamsFromContext)
}
