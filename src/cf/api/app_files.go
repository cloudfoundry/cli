package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
)

type AppFilesRepository interface {
	ListFiles(app cf.Application, path string) (files string, apiErr *net.ApiError)
}

type CloudControllerAppFilesRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerAppFilesRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerAppFilesRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppFilesRepository) ListFiles(app cf.Application, path string) (files string, apiErr *net.ApiError) {
	url := fmt.Sprintf("%s/v2/apps/%s/instances/0/files/%s", repo.config.Target, app.Guid, path)
	request, apiErr := repo.gateway.NewRequest("GET", url, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	files, _, apiErr = repo.gateway.PerformRequestForTextResponse(request)
	return
}
