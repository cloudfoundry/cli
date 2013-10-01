package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeAppFilesRepo struct{
	Application cf.Application
	Path string
	FileList string
}


func (repo *FakeAppFilesRepo)ListFiles(app cf.Application, path string) (files string, apiStatus net.ApiStatus) {
	repo.Application = app
	repo.Path = path

	files = repo.FileList

	return
}
