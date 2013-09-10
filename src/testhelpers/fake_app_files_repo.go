package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeAppFilesRepo struct{
	Application cf.Application
	Path string
	FileList string
}


func (repo *FakeAppFilesRepo)ListFiles(app cf.Application, path string) (files string, apiErr *api.ApiError) {
	repo.Application = app
	repo.Path = path

	files = repo.FileList

	return
}
