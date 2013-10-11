package api

import (
	"cf"
	"cf/net"
)

type FakeApplicationBitsRepository struct {
	UploadedApp cf.Application
	UploadedDir string
	UploadAppErr bool
}

func (repo *FakeApplicationBitsRepository) UploadApp(app cf.Application, dir string) (apiResponse net.ApiResponse) {
	repo.UploadedDir = dir
	repo.UploadedApp = app

	if repo.UploadAppErr {
		apiResponse = net.NewApiResponseWithMessage("Error uploading app")
	}

	return
}
