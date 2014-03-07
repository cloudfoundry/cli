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
	uaaEndpoint := repo.config.AuthenticationEndpoint()
	if uaaEndpoint == "" {
		return errors.NewErrorWithMessage("Authorization endpoint missing from config file")
	}

	url := fmt.Sprintf("%s/Users/%s/password", uaaEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)

	return repo.gateway.UpdateResource(url, repo.config.AccessToken(), strings.NewReader(body))
}
