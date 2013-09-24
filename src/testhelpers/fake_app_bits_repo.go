package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeApplicationBitsRepository struct {
	UploadedApp cf.Application
	UploadedDir string
}

func (repo *FakeApplicationBitsRepository) UploadApp(app cf.Application, dir string) (apiErr *net.ApiError) {
	repo.UploadedDir = dir
	repo.UploadedApp = app

	return
}
