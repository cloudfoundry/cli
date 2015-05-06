package actors

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/api/application_bits"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/gofileutils/fileutils"
)

type PushActor interface {
	UploadApp(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) error
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

func (actor PushActorImpl) GatherFiles(appDir string, uploadDir string) (presentFiles []resources.AppFileResource, hasFileToUpload bool, apiErr error) {
	if actor.zipper.IsZipFile(appDir) {
		fileutils.TempDir("unzipped-app", func(tmpDir string, err error) {
			err = actor.zipper.Unzip(appDir, tmpDir)
			if err != nil {
				presentFiles = nil
				apiErr = err
				return
			}
			presentFiles, hasFileToUpload, apiErr = actor.copyUploadableFiles(tmpDir, uploadDir)
		})
	} else {
		presentFiles, hasFileToUpload, apiErr = actor.copyUploadableFiles(appDir, uploadDir)
	}
	return presentFiles, hasFileToUpload, apiErr
}

func (actor PushActorImpl) UploadApp(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) error {
	return actor.appBitsRepo.UploadBits(appGuid, zipFile, presentFiles)
}

func (actor PushActorImpl) copyUploadableFiles(appDir string, uploadDir string) (presentFiles []resources.AppFileResource, hasFileToUpload bool, err error) {
	// Find which files need to be uploaded
	allAppFiles, err := actor.appfiles.AppFilesInDir(appDir)
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
	fileutils.CopyPathToPath(filepath.Join(appDir, ".cfignore"), filepath.Join(uploadDir, ".cfignore")) //error handling?

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
