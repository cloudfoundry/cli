package api

import (
	"github.com/cloudfoundry/cli/cf/errors"
)

type FakeApplicationBitsRepository struct {
	UploadedAppGuid string
	UploadedDir     string
	UploadAppErr    bool

	CallbackPath      string
	CallbackZipSize   uint64
	CallbackFileCount uint64
}

func (repo *FakeApplicationBitsRepository) UploadApp(appGuid, dir string, cb func(path string, zipSize, fileCount uint64)) (apiErr error) {
	repo.UploadedDir = dir
	repo.UploadedAppGuid = appGuid

	if repo.UploadAppErr {
		apiErr = errors.New("Error uploading app")
		return
	}

	cb(repo.CallbackPath, repo.CallbackZipSize, repo.CallbackFileCount)

	return
}
