package api

import (
	"cf"
	"cf/configuration"
	"fmt"
)

type AppFilesRepository interface {
	ListFiles(app cf.Application, path string) (files string, apiErr *ApiError)
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

func (repo CloudControllerAppFilesRepository) ListFiles(app cf.Application, path string) (files string, apiErr *ApiError) {
	url := fmt.Sprintf("%s/v2/apps/%s/instances/0/files/%s", repo.config.Target, app.Guid, path)
	request, apiErr := NewRequest("GET", url, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	files, apiErr = repo.client.PerformRequestForTextResponse(request)
	return
}
