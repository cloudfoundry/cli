package fakes

type FakeAppFilesRepo struct {
	AppGuid  string
	Instance int
	Path     string
	FileList string
}

func (repo *FakeAppFilesRepo) ListFiles(appGuid string, instance int, path string) (files string, apiErr error) {
	repo.AppGuid = appGuid
	repo.Instance = instance
	repo.Path = path

	files = repo.FileList

	return
}
