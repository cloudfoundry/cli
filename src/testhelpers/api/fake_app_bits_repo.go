package api

import (
	"cf/net"
)

type FakeApplicationBitsRepository struct {
	UploadedAppGuid string
	UploadedDir     string
	UploadAppErr    bool

	CallbackPath   	string
	CallbackZipSize   uint64
	CallbackFileCount uint64
}

func (repo *FakeApplicationBitsRepository) UploadApp(appGuid, dir string, cb func(path string,zipSize, fileCount uint64)) (apiResponse net.ApiResponse) {
	repo.UploadedDir = dir
	repo.UploadedAppGuid = appGuid

	if repo.UploadAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error uploading app")
		return
	}

	cb(repo.CallbackPath,repo.CallbackZipSize, repo.CallbackFileCount)

	return
}
