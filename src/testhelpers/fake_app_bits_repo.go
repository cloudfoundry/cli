package testhelpers

import (
	"cf"
	"bytes"
	"cf/net"
	"os"
)

type FakeApplicationBitsRepository struct {
	UploadedApp cf.Application
	UploadedZipBuffer *bytes.Buffer

	CreateUploadDirApp cf.Application
	CreateUploadDirAppDir string

	GetFilesToUploadApp cf.Application
	GetFilesToUploadAllAppFiles []cf.AppFile
}

func (repo *FakeApplicationBitsRepository) Upload(app cf.Application, zipBuffer *bytes.Buffer) (apiErr *net.ApiError) {
	repo.UploadedZipBuffer = zipBuffer
	repo.UploadedApp = app

	return
}

func (repo *FakeApplicationBitsRepository) CreateUploadDir(app cf.Application, appDir string) (uploadDir string, err error) {
	repo.CreateUploadDirApp = app
	repo.CreateUploadDirAppDir = appDir

	uploadDir, err  = os.Getwd()
	return
}

func (repo *FakeApplicationBitsRepository) GetFilesToUpload(app cf.Application, allAppFiles []cf.AppFile) (appFilesToUpload []cf.AppFile, err error) {
	repo.GetFilesToUploadApp = app
	repo.GetFilesToUploadAllAppFiles = allAppFiles
	return
}
