package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/net"
	"fmt"
	"strings"
)

type PasswordRepository interface {
	UpdatePassword(old string, new string) errors.Error
}

type CloudControllerPasswordRepository struct {
	config       configuration.Reader
	gateway      net.Gateway
	endpointRepo EndpointRepository
}

func NewCloudControllerPasswordRepository(config configuration.Reader, gateway net.Gateway, endpointRepo EndpointRepository) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.endpointRepo = endpointRepo
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) (apiResponse errors.Error) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetUAAEndpoint()
	if apiResponse != nil {
		return
	}

	path := fmt.Sprintf("%s/Users/%s/password", uaaEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)

	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}
