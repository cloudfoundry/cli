package api

import (
	"cf/configuration"
	"cf/net"
	"fmt"
)

type AppFilesRepository interface {
	ListFiles(appGuid, path string) (files string, apiResponse net.ApiResponse)
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

func (repo CloudControllerAppFilesRepository) ListFiles(appGuid, path string) (files string, apiResponse net.ApiResponse) {
	url := fmt.Sprintf("%s/v2/apps/%s/instances/0/files/%s", repo.config.Target, appGuid, path)
	request, apiResponse := repo.gateway.NewRequest("GET", url, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	files, _, apiResponse = repo.gateway.PerformRequestForTextResponse(request)
	return
}
