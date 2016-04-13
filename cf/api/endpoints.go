package api

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/net"
)

type RemoteInfoRepository struct {
	gateway net.Gateway
}

func NewEndpointRepository(gateway net.Gateway) RemoteInfoRepository {
	r := RemoteInfoRepository{
		gateway: gateway,
	}
	return r
}

func (repo RemoteInfoRepository) GetCCInfo(endpoint string) (*core_config.CCInfo, string, error) {
	if strings.HasPrefix(endpoint, "http") {
		serverResponse, err := repo.getCCAPIInfo(endpoint)
		if err != nil {
			return nil, "", err
		}

		return serverResponse, endpoint, nil
	}

	finalEndpoint := "https://" + endpoint
	serverResponse, err := repo.getCCAPIInfo(finalEndpoint)
	if err != nil {
		if _, ok := err.(*errors.InvalidSSLCert); ok {
			return nil, "", err
		}

		finalEndpoint = "http://" + endpoint
		serverResponse, err = repo.getCCAPIInfo(finalEndpoint)
		if err != nil {
			return nil, "", err
		}
	}

	return serverResponse, finalEndpoint, nil
}

func (repo RemoteInfoRepository) getCCAPIInfo(endpoint string) (*core_config.CCInfo, error) {
	serverResponse := new(core_config.CCInfo)
	err := repo.gateway.GetResource(endpoint+"/v2/info", &serverResponse)
	if err != nil {
		return nil, err
	}

	return serverResponse, nil
}
