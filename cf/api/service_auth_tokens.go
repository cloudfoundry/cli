package api

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"net/url"
	"strings"
)

type ServiceAuthTokenRepository interface {
	FindAll() (authTokens []models.ServiceAuthTokenFields, apiErr error)
	FindByLabelAndProvider(label, provider string) (authToken models.ServiceAuthTokenFields, apiErr error)
	Create(authToken models.ServiceAuthTokenFields) (apiErr error)
	Update(authToken models.ServiceAuthTokenFields) (apiErr error)
	Delete(authToken models.ServiceAuthTokenFields) (apiErr error)
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

func (repo CloudControllerServiceAuthTokenRepository) FindAll() (authTokens []models.ServiceAuthTokenFields, apiErr error) {
	return repo.findAllWithPath("/v2/service_auth_tokens")
}

func (repo CloudControllerServiceAuthTokenRepository) FindByLabelAndProvider(label, provider string) (authToken models.ServiceAuthTokenFields, apiErr error) {
	path := fmt.Sprintf("/v2/service_auth_tokens?q=%s", url.QueryEscape("label:"+label+";provider:"+provider))
	authTokens, apiErr := repo.findAllWithPath(path)
	if apiErr != nil {
		return
	}

	if len(authTokens) == 0 {
		apiErr = errors.NewModelNotFoundError("Service Auth Token", label+" "+provider)
		return
	}

	authToken = authTokens[0]
	return
}

func (repo CloudControllerServiceAuthTokenRepository) findAllWithPath(path string) ([]models.ServiceAuthTokenFields, error) {
	var authTokens []models.ServiceAuthTokenFields
	apiErr := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.AuthTokenResource{},
		func(resource interface{}) bool {
			if at, ok := resource.(resources.AuthTokenResource); ok {
				authTokens = append(authTokens, at.ToFields())
			}
			return true
		})

	return authTokens, apiErr
}

func (repo CloudControllerServiceAuthTokenRepository) Create(authToken models.ServiceAuthTokenFields) (apiErr error) {
	body := fmt.Sprintf(`{"label":"%s","provider":"%s","token":"%s"}`, authToken.Label, authToken.Provider, authToken.Token)
	path := fmt.Sprintf("%s/v2/service_auth_tokens", repo.config.ApiEndpoint())
	return repo.gateway.CreateResource(path, strings.NewReader(body))
}

func (repo CloudControllerServiceAuthTokenRepository) Delete(authToken models.ServiceAuthTokenFields) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_auth_tokens/%s", repo.config.ApiEndpoint(), authToken.Guid)
	return repo.gateway.DeleteResource(path)
}

func (repo CloudControllerServiceAuthTokenRepository) Update(authToken models.ServiceAuthTokenFields) (apiErr error) {
	body := fmt.Sprintf(`{"token":"%s"}`, authToken.Token)
	path := fmt.Sprintf("%s/v2/service_auth_tokens/%s", repo.config.ApiEndpoint(), authToken.Guid)
	return repo.gateway.UpdateResource(path, strings.NewReader(body))
}
