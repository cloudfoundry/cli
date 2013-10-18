package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
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
	FindAll() (authTokens []cf.ServiceAuthToken, apiResponse net.ApiResponse)
	FindByLabelAndProvider(label, provider string) (authToken cf.ServiceAuthToken, apiResponse net.ApiResponse)
	Create(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse)
	Update(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse)
	Delete(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse)
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
	resources := new(PaginatedAuthTokenResources)

	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
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

func (repo CloudControllerServiceAuthTokenRepository) Create(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(`{"label":"%s","provider":"%s","token":"%s"}`, authToken.Label, authToken.Provider, authToken.Token)
	path := fmt.Sprintf("%s/v2/service_auth_tokens", repo.config.Target)
	return repo.gateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerServiceAuthTokenRepository) Delete(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens/%s", repo.config.Target, authToken.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerServiceAuthTokenRepository) Update(authToken cf.ServiceAuthToken) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(`{"token":"%s"}`, authToken.Token)
	path := fmt.Sprintf("%s/v2/service_auth_tokens/%s", repo.config.Target, authToken.Guid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}
