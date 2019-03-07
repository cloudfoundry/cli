package api

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/net"
)

//go:generate counterfeiter . ClientRepository

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
