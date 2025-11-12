package api

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/v8/cf/api/resources"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/errors"
	. "code.cloudfoundry.org/cli/v8/cf/i18n"
	"code.cloudfoundry.org/cli/v8/cf/net"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ClientRepository

type ClientRepository interface {
	ClientExists(clientID string) (exists bool, apiErr error)
}

type CloudControllerClientRepository struct {
	config     coreconfig.Reader
	uaaGateway net.Gateway
}

func NewCloudControllerClientRepository(config coreconfig.Reader, uaaGateway net.Gateway) (repo CloudControllerClientRepository) {
	repo.config = config
	repo.uaaGateway = uaaGateway
	return
}

func (repo CloudControllerClientRepository) ClientExists(clientID string) (exists bool, apiErr error) {
	exists = false
	uaaEndpoint, apiErr := repo.getAuthEndpoint()
	if apiErr != nil {
		return exists, apiErr
	}

	path := fmt.Sprintf("%s/oauth/clients/%s", uaaEndpoint, clientID)

	uaaResponse := new(resources.UAAUserResources)
	apiErr = repo.uaaGateway.GetResource(path, uaaResponse)
	if apiErr != nil {
		if errType, ok := apiErr.(errors.HTTPError); ok {
			switch errType.StatusCode() {
			case http.StatusNotFound:
				return false, errors.NewModelNotFoundError("Client", clientID)
			case http.StatusForbidden:
				return false, errors.NewAccessDeniedError()
			}
		}
		return false, apiErr
	}
	return true, nil
}

func (repo CloudControllerClientRepository) getAuthEndpoint() (string, error) {
	uaaEndpoint := repo.config.UaaEndpoint()
	if uaaEndpoint == "" {
		return "", errors.New(T("UAA endpoint missing from config file"))
	}
	return uaaEndpoint, nil
}
