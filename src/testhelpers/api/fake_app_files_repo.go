package api

import (
	"cf/errors"
)

type FakeAppFilesRepo struct {
	AppGuid  string
	Path     string
	FileList string
}

func (repo *FakeAppFilesRepo) ListFiles(appGuid, path string) (files string, apiResponse errors.Error) {
	repo.AppGuid = appGuid
	repo.Path = path

	files = repo.FileList

	return
}
