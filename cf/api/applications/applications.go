package applications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

//go:generate counterfeiter -o fakes/fake_application_repository.go . ApplicationRepository
type ApplicationRepository interface {
	Create(params models.AppParams) (createdApp models.Application, apiErr error)
	GetApp(appGuid string) (models.Application, error)
	Read(name string) (app models.Application, apiErr error)
	ReadFromSpace(name string, spaceGuid string) (app models.Application, apiErr error)
	Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiErr error)
	Delete(appGuid string) (apiErr error)
	ReadEnv(guid string) (*models.Environment, error)
	CreateRestageRequest(guid string) (apiErr error)
}

type CloudControllerApplicationRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerApplicationRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerApplicationRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerApplicationRepository) Create(params models.AppParams) (createdApp models.Application, apiErr error) {
	data, err := repo.formatAppJSON(params)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", T("Failed to marshal JSON"), err.Error())
		return
	}

	resource := new(resources.ApplicationResource)
	apiErr = repo.gateway.CreateResource(repo.config.ApiEndpoint(), "/v2/apps", bytes.NewReader(data), resource)
	if apiErr != nil {
		return
	}

	createdApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) GetApp(appGuid string) (app models.Application, apiErr error) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.ApiEndpoint(), appGuid)
	appResources := new(resources.ApplicationResource)

	apiErr = repo.gateway.GetResource(path, appResources)
	if apiErr != nil {
		return
	}

	app = appResources.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Read(name string) (app models.Application, apiErr error) {
	return repo.ReadFromSpace(name, repo.config.SpaceFields().Guid)
}

func (repo CloudControllerApplicationRepository) ReadFromSpace(name string, spaceGuid string) (app models.Application, apiErr error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=%s&inline-relations-depth=1", repo.config.ApiEndpoint(), spaceGuid, url.QueryEscape("name:"+name))
	appResources := new(resources.PaginatedApplicationResources)
	apiErr = repo.gateway.GetResource(path, appResources)
	if apiErr != nil {
		return
	}

	if len(appResources.Resources) == 0 {
		apiErr = errors.NewModelNotFoundError("App", name)
		return
	}

	res := appResources.Resources[0]
	app = res.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) Update(appGuid string, params models.AppParams) (updatedApp models.Application, apiErr error) {
	data, err := repo.formatAppJSON(params)
	if err != nil {
		apiErr = fmt.Errorf("%s: %s", T("Failed to marshal JSON"), err.Error())
		return
	}

	path := fmt.Sprintf("/v2/apps/%s?inline-relations-depth=1", appGuid)
	resource := new(resources.ApplicationResource)
	apiErr = repo.gateway.UpdateResource(repo.config.ApiEndpoint(), path, bytes.NewReader(data), resource)
	if apiErr != nil {
		return
	}

	updatedApp = resource.ToModel()
	return
}

func (repo CloudControllerApplicationRepository) formatAppJSON(input models.AppParams) ([]byte, error) {
	appResource := resources.NewApplicationEntityFromAppParams(input)
	data, err := json.Marshal(appResource)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (repo CloudControllerApplicationRepository) Delete(appGuid string) (apiErr error) {
	path := fmt.Sprintf("/v2/apps/%s?recursive=true", appGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}

func (repo CloudControllerApplicationRepository) ReadEnv(guid string) (*models.Environment, error) {
	var (
		err error
	)

	path := fmt.Sprintf("%s/v2/apps/%s/env", repo.config.ApiEndpoint(), guid)
	appResource := models.NewEnvironment()

	err = repo.gateway.GetResource(path, appResource)
	if err != nil {
		return &models.Environment{}, err
	}

	return appResource, err
}

func (repo CloudControllerApplicationRepository) CreateRestageRequest(guid string) error {
	path := fmt.Sprintf("/v2/apps/%s/restage", guid)
	return repo.gateway.CreateResource(repo.config.ApiEndpoint(), path, strings.NewReader(""), nil)
}
