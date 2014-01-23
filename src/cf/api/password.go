package api

import (
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type PasswordRepository interface {
	UpdatePassword(old string, new string) net.ApiResponse
}

type CloudControllerPasswordRepository struct {
	config       *configuration.Configuration
	gateway      net.Gateway
	endpointRepo EndpointRepository
}

func NewCloudControllerPasswordRepository(config *configuration.Configuration, gateway net.Gateway, endpointRepo EndpointRepository) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.endpointRepo = endpointRepo
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) (apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetUAAEndpoint()
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/Users/%s/password", uaaEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)

	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}
