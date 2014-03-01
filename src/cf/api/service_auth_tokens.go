package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
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
	FindAll() (authTokens []models.ServiceAuthTokenFields, apiErr errors.Error)
	FindByLabelAndProvider(label, provider string) (authToken models.ServiceAuthTokenFields, apiErr errors.Error)
	Create(authToken models.ServiceAuthTokenFields) (apiErr errors.Error)
	Update(authToken models.ServiceAuthTokenFields) (apiErr errors.Error)
	Delete(authToken models.ServiceAuthTokenFields) (apiErr errors.Error)
}

type CloudControllerServiceAuthTokenRepository struct {
	gateway net.Gateway
	config  configuration.Reader
}

func NewCloudControllerServiceAuthTokenRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerServiceAuthTokenRepository) {
	repo.gateway = gateway
	repo.config = config
	return
}

func (repo CloudControllerServiceAuthTokenRepository) FindAll() (authTokens []models.ServiceAuthTokenFields, apiErr errors.Error) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens", repo.config.ApiEndpoint())
	return repo.findAllWithPath(path)
}

func (repo CloudControllerServiceAuthTokenRepository) FindByLabelAndProvider(label, provider string) (authToken models.ServiceAuthTokenFields, apiErr errors.Error) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens?q=%s", repo.config.ApiEndpoint(), url.QueryEscape("label:"+label+";provider:"+provider))
	authTokens, apiErr := repo.findAllWithPath(path)
	if apiErr != nil {
		return
	}

	if len(authTokens) == 0 {
		apiErr = errors.NewNotFoundError("Service Auth Token %s %s not found", label, provider)
		return
	}

	authToken = authTokens[0]
	return
}

func (repo CloudControllerServiceAuthTokenRepository) findAllWithPath(path string) (authTokens []models.ServiceAuthTokenFields, apiErr errors.Error) {
	resources := new(PaginatedAuthTokenResources)

	apiErr = repo.gateway.GetResource(path, repo.config.AccessToken(), resources)
	if apiErr != nil {
		return
	}

	for _, resource := range resources.Resources {
		authTokens = append(authTokens, models.ServiceAuthTokenFields{
			Guid:     resource.Metadata.Guid,
			Label:    resource.Entity.Label,
			Provider: resource.Entity.Provider,
		})
	}
	return
}

func (repo CloudControllerServiceAuthTokenRepository) Create(authToken models.ServiceAuthTokenFields) (apiErr errors.Error) {
	body := fmt.Sprintf(`{"label":"%s","provider":"%s","token":"%s"}`, authToken.Label, authToken.Provider, authToken.Token)
	path := fmt.Sprintf("%s/v2/service_auth_tokens", repo.config.ApiEndpoint())
	return repo.gateway.CreateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}

func (repo CloudControllerServiceAuthTokenRepository) Delete(authToken models.ServiceAuthTokenFields) (apiErr errors.Error) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens/%s", repo.config.ApiEndpoint(), authToken.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken())
}

func (repo CloudControllerServiceAuthTokenRepository) Update(authToken models.ServiceAuthTokenFields) (apiErr errors.Error) {
	body := fmt.Sprintf(`{"token":"%s"}`, authToken.Token)
	path := fmt.Sprintf("%s/v2/service_auth_tokens/%s", repo.config.ApiEndpoint(), authToken.Guid)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}
