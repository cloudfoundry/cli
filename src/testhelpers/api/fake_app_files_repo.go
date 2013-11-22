package api

import (
	"cf/net"
)

type FakeAppFilesRepo struct{
	AppGuid string
	Path string
	FileList string
}


func (repo *FakeAppFilesRepo)ListFiles(appGuid, path string) (files string, apiResponse net.ApiResponse) {
	repo.AppGuid= appGuid
	repo.Path = path

	files = repo.FileList

	return
}
