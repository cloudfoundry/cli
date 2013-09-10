package api

import (
	"cf"
	"cf/configuration"
	"fmt"
)

type AppFilesRepository interface {
	ListFiles(app cf.Application, path string) (files string, err error)
}

type CloudControllerAppFilesRepository struct {
	config *configuration.Configuration
	client ApiClient
}

func NewCloudControllerAppFilesRepository(config *configuration.Configuration, client ApiClient) (repo CloudControllerAppFilesRepository) {
	repo.config = config
	repo.client = client
	return
}

func (repo CloudControllerAppFilesRepository) ListFiles(app cf.Application, path string) (files string, err error) {
	url := fmt.Sprintf("%s/v2/apps/%s/instances/0/files/%s", repo.config.Target, app.Guid, path)
	request, err := NewRequest("GET", url, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	files, err = repo.client.PerformRequestForTextResponse(request)
	return
}
