package testhelpers

import (
	"cf"
	"bytes"
	"cf/net"
)

type FakeApplicationBitsRepository struct {
	UploadedApp cf.Application
	UploadedZipBuffer *bytes.Buffer
}

func (repo *FakeApplicationBitsRepository) Upload(app cf.Application, zipBuffer *bytes.Buffer) (apiErr *net.ApiError) {
	repo.UploadedZipBuffer = zipBuffer
	repo.UploadedApp = app

	return
}
