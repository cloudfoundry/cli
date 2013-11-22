package api

import (
	"cf/net"
)

type FakeApplicationBitsRepository struct {
	UploadedAppGuid string
	UploadedDir string
	UploadAppErr bool
}

func (repo *FakeApplicationBitsRepository) UploadApp(appGuid, dir string) (apiResponse net.ApiResponse) {
	repo.UploadedDir = dir
	repo.UploadedAppGuid = appGuid

	if repo.UploadAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error uploading app")
	}

	return
}
