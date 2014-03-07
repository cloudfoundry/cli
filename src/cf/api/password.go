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
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerPasswordRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) errors.Error {
	uaaEndpoint := repo.config.AuthorizationEndpoint()
	if uaaEndpoint == "" {
		return errors.NewErrorWithMessage("UAA endpoint missing from config file")
	}

	path := fmt.Sprintf("%s/Users/%s/password", uaaEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)

	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}
