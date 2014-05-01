package api

type FakeAppFilesRepo struct {
	AppGuid  string
	Path     string
	FileList string
}

func (repo *FakeAppFilesRepo) ListFiles(appGuid, path string) (files string, apiErr error) {
	repo.AppGuid = appGuid
	repo.Path = path

	files = repo.FileList

	return
}
