package testhelpers

import "cf"

type FakeAppFilesRepo struct{
	Application cf.Application
	Path string
	FileList string
}


func (repo *FakeAppFilesRepo)ListFiles(app cf.Application, path string) (files string, err error) {
	repo.Application = app
	repo.Path = path

	files = repo.FileList

	return
}
