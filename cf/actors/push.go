package actors

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/api/application_bits"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/gofileutils/fileutils"
)

//go:generate counterfeiter -o fakes/fake_push_actor.go . PushActor
type PushActor interface {
	UploadApp(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) error
	PopulateFileMode(appDir string, presentFiles []resources.AppFileResource) ([]resources.AppFileResource, error)
	GatherFiles(appDir string, uploadDir string) ([]resources.AppFileResource, bool, error)
}

type PushActorImpl struct {
	appBitsRepo application_bits.ApplicationBitsRepository
	appfiles    app_files.AppFiles
	zipper      app_files.Zipper
}

func NewPushActor(appBitsRepo application_bits.ApplicationBitsRepository, zipper app_files.Zipper, appfiles app_files.AppFiles) PushActor {
	return PushActorImpl{
		appBitsRepo: appBitsRepo,
		appfiles:    appfiles,
		zipper:      zipper,
	}
}

func (actor PushActorImpl) PopulateFileMode(appDir string, presentFiles []resources.AppFileResource) ([]resources.AppFileResource, error) {
	for i := range presentFiles {
		fileInfo, err := os.Lstat(filepath.Join(appDir, presentFiles[i].Path))
		if err != nil {
			return presentFiles, err
		}
		presentFiles[i].Mode = fmt.Sprintf("%#o", fileInfo.Mode())
	}

	return presentFiles, nil
}

func (actor PushActorImpl) GatherFiles(appDir string, uploadDir string) ([]resources.AppFileResource, bool, error) {
	if actor.zipper.IsZipFile(appDir) {
		tmpDir, err := ioutil.TempDir("", "unzipped-app")
		if err != nil {
			return []resources.AppFileResource{}, false, err
		}
		defer os.RemoveAll(tmpDir)

		err = actor.zipper.Unzip(appDir, tmpDir)
		if err != nil {
			return []resources.AppFileResource{}, false, err
		}

		appDir = tmpDir
	}
	appDir, err := filepath.EvalSymlinks(appDir)
	if err != nil {
		return []resources.AppFileResource{}, false, err
	}

	files, hasFileToUpload, err := actor.copyUploadableFiles(appDir, uploadDir)
	if err != nil {
		return []resources.AppFileResource{}, false, err
	}

	filesWithFileMode, err := actor.PopulateFileMode(appDir, files)
	if err != nil {
		return []resources.AppFileResource{}, false, err
	}

	return filesWithFileMode, hasFileToUpload, nil
}

func (actor PushActorImpl) UploadApp(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) error {
	return actor.appBitsRepo.UploadBits(appGuid, zipFile, presentFiles)
}

func (actor PushActorImpl) copyUploadableFiles(appDir string, uploadDir string) (presentFiles []resources.AppFileResource, hasFileToUpload bool, err error) {
	// Find which files need to be uploaded
	var allAppFiles []models.AppFileFields
	allAppFiles, err = actor.appfiles.AppFilesInDir(appDir)
	if err != nil {
		return
	}

	appFilesToUpload, presentFiles, apiErr := actor.getFilesToUpload(allAppFiles)
	if apiErr != nil {
		err = errors.New(apiErr.Error())
		return
	}
	hasFileToUpload = len(appFilesToUpload) > 0

	// Copy files into a temporary directory and return it
	err = actor.appfiles.CopyFiles(appFilesToUpload, appDir, uploadDir)
	if err != nil {
		return
	}

	// copy cfignore if present
	// TODO: move this into AppFiles, see: https://github.com/cloudfoundry/cli/commit/dcf7dfe5305a10b569e7a2bfb2ed3db25096065c
	fileutils.CopyPathToPath(filepath.Join(appDir, ".cfignore"), filepath.Join(uploadDir, ".cfignore")) // error handling?

	return
}

func (actor PushActorImpl) getFilesToUpload(allAppFiles []models.AppFileFields) (appFilesToUpload []models.AppFileFields, presentFiles []resources.AppFileResource, apiErr error) {
	appFilesRequest := []resources.AppFileResource{}
	for _, file := range allAppFiles {
		appFilesRequest = append(appFilesRequest, resources.AppFileResource{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		})
	}

	presentFiles, apiErr = actor.appBitsRepo.GetApplicationFiles(appFilesRequest)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	appFilesToUpload = make([]models.AppFileFields, len(allAppFiles))
	copy(appFilesToUpload, allAppFiles)
	for _, file := range presentFiles {
		appFile := models.AppFileFields{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		}
		appFilesToUpload = actor.deleteAppFile(appFilesToUpload, appFile)
	}

	return
}

func (actor PushActorImpl) deleteAppFile(appFiles []models.AppFileFields, targetFile models.AppFileFields) []models.AppFileFields {
	for i, file := range appFiles {
		if file.Path == targetFile.Path {
			appFiles[i] = appFiles[len(appFiles)-1]
			return appFiles[:len(appFiles)-1]
		}
	}
	return appFiles
}
