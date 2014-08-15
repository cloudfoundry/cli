package fakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
)

type FakeApplicationBitsRepository struct {
	UploadedAppGuid string
	UploadedDir     string
	UploadAppErr    bool

	CallbackPath      string
	CallbackZipSize   int64
	CallbackFileCount int64
}

func (repo *FakeApplicationBitsRepository) UploadApp(appGuid, dir string, cb func(path string, zipSize, fileCount int64)) (apiErr error) {
	repo.UploadedDir = dir
	repo.UploadedAppGuid = appGuid

	if repo.UploadAppErr {
		apiErr = errors.New("Error uploading app")
		return
	}

	cb(repo.CallbackPath, repo.CallbackZipSize, repo.CallbackFileCount)

	return
}
