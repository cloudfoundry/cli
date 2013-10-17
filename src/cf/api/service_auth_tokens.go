package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"io"
	"strings"
)

type PaginatedAuthTokenResources struct {
	Resources []AuthTokenResource
}

type AuthTokenResource struct {
	Resource
	Entity AuthTokenEntity
}

type AuthTokenEntity struct {
	Label    string
	Provider string
}

type ServiceAuthTokenRepository interface {
	Create(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse)
	Update(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse)
	Delete(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse)
	FindAll() (authTokens []cf.ServiceAuthToken, apiResponse net.ApiResponse)
	FindByLabelAndProvider(label, provider string) (authToken cf.ServiceAuthToken, apiResponse net.ApiResponse)
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

func (repo CloudControllerServiceAuthTokenRepository) FindAll() (authTokens []cf.ServiceAuthToken, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens", repo.config.Target)
	return repo.findAllWithPath(path)
}

func (repo CloudControllerServiceAuthTokenRepository) FindByLabelAndProvider(label, provider string) (authToken cf.ServiceAuthToken, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens?q=label:%s;provider:%s", repo.config.Target, label, provider)
	authTokens, apiResponse := repo.findAllWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(authTokens) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Service Auth Token %s %s not found", label, provider)
		return
	}

	authToken = authTokens[0]
	return
}

func (repo CloudControllerServiceAuthTokenRepository) findAllWithPath(path string) (authTokens []cf.ServiceAuthToken, apiResponse net.ApiResponse) {
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedAuthTokenResources)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, resource := range resources.Resources {
		authTokens = append(authTokens, cf.ServiceAuthToken{
			Guid:     resource.Metadata.Guid,
			Label:    resource.Entity.Label,
			Provider: resource.Entity.Provider,
		})
	}
	return
}

func (repo CloudControllerServiceAuthTokenRepository) Delete(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	return repo.updateOrDelete(authToken, "DELETE", nil)
}

func (repo CloudControllerServiceAuthTokenRepository) Update(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(`{"token":"%s"}`, authToken.Token)
	return repo.updateOrDelete(authToken, "PUT", strings.NewReader(body))
}

func (repo CloudControllerServiceAuthTokenRepository) updateOrDelete(authToken cf.ServiceAuthToken, verb string, body io.Reader) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens/%s", repo.config.Target, authToken.Guid)

	request, apiResponse := repo.gateway.NewRequest(verb, path, repo.config.AccessToken, body)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
