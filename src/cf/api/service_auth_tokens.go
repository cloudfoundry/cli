package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type ServiceAuthTokenRepository interface {
	Create(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse)
}

type CloudControllerServiceAuthTokenRepository struct {
	gateway net.Gateway
	config  *configuration.Configuration
}

func NewCloudControllerServiceAuthTokenRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerServiceAuthTokenRepository) {
	repo.gateway = gateway
	repo.config = config
	return
}

func (repo CloudControllerServiceAuthTokenRepository) Create(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens", repo.config.Target)
	body := fmt.Sprintf(`{"label":"%s","provider":"%s","token":"%s"}`, authToken.Label, authToken.Provider, authToken.Token)

	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
