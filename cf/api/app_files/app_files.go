package app_files

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
)

type AppFilesRepository interface {
	ListFiles(appGuid string, instance int, path string) (files string, apiErr error)
}

type CloudControllerAppFilesRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerAppFilesRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerAppFilesRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppFilesRepository) ListFiles(appGuid string, instance int, path string) (files string, apiErr error) {
	url := fmt.Sprintf("%s/v2/apps/%s/instances/%d/files/%s", repo.config.ApiEndpoint(), appGuid, instance, path)
	request, apiErr := repo.gateway.NewRequest("GET", url, repo.config.AccessToken(), nil)
	if apiErr != nil {
		return
	}

	files, _, apiErr = repo.gateway.PerformRequestForTextResponse(request)
	return
}
