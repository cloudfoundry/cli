package actors

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/cli/cf/api/application_bits"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/gofileutils/fileutils"
)

//go:generate counterfeiter -o fakes/fake_push_actor.go . PushActor
type PushActor interface {
	UploadApp(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) error
	ProcessPath(dirOrZipFile string, f func(string)) error
	GatherFiles(localFiles []models.AppFileFields, appDir string, uploadDir string) ([]resources.AppFileResource, bool, error)
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

func (actor PushActorImpl) ProcessPath(dirOrZipFile string, f func(string)) error {
	if !actor.zipper.IsZipFile(dirOrZipFile) {
		appDir, err := filepath.EvalSymlinks(dirOrZipFile)
		if err == nil {
			f(appDir)
		}
		return err
	}

	tempDir, err := ioutil.TempDir("", "unzipped-app")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	err = actor.zipper.Unzip(dirOrZipFile, tempDir)
	if err != nil {
		return err
	}

	f(tempDir)

	return nil
}

func (actor PushActorImpl) GatherFiles(localFiles []models.AppFileFields, appDir string, uploadDir string) ([]resources.AppFileResource, bool, error) {
	appFileResource := []resources.AppFileResource{}
	for _, file := range localFiles {
		appFileResource = append(appFileResource, resources.AppFileResource{
			Path: file.Path,
			Sha1: file.Sha1,
			Size: file.Size,
		})
	}

	remoteFiles, err := actor.appBitsRepo.GetApplicationFiles(appFileResource)
	if err != nil {
		return []resources.AppFileResource{}, false, err
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
		fileInfo, err := os.Lstat(filepath.Join(appDir, remoteFiles[i].Path))
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

func (actor PushActorImpl) UploadApp(appGuid string, zipFile *os.File, presentFiles []resources.AppFileResource) error {
	return actor.appBitsRepo.UploadBits(appGuid, zipFile, presentFiles)
}
