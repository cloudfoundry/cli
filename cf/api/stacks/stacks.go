package stacks

import (
	"fmt"
	"net/url"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type StackRepository interface {
	FindByName(name string) (stack models.Stack, apiErr error)
	FindByGUID(guid string) (models.Stack, error)
	FindAll() (stacks []models.Stack, apiErr error)
}

type CloudControllerStackRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerStackRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerStackRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerStackRepository) FindByGUID(guid string) (models.Stack, error) {
	stackRequest := resources.StackResource{}
	path := fmt.Sprintf("%s/v2/stacks/%s", repo.config.ApiEndpoint(), guid)
	err := repo.gateway.GetResource(path, &stackRequest)
	if err != nil {
		if errNotFound, ok := err.(*errors.HttpNotFoundError); ok {
			return models.Stack{}, errNotFound
		}

		return models.Stack{}, fmt.Errorf(T("Error retrieving stacks: {{.Error}}", map[string]interface{}{
			"Error": err.Error(),
		}))
	}

	return *stackRequest.ToFields(), nil
}

func (repo CloudControllerStackRepository) FindByName(name string) (stack models.Stack, apiErr error) {
	path := fmt.Sprintf("/v2/stacks?q=%s", url.QueryEscape("name:"+name))
	stacks, apiErr := repo.findAllWithPath(path)
	if apiErr != nil {
		return
	}

	if len(stacks) == 0 {
		apiErr = errors.NewModelNotFoundError("Stack", name)
		return
	}

	stack = stacks[0]
	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []models.Stack, apiErr error) {
	return repo.findAllWithPath("/v2/stacks")
}

func (repo CloudControllerStackRepository) findAllWithPath(path string) ([]models.Stack, error) {
	var stacks []models.Stack
	apiErr := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.StackResource{},
		func(resource interface{}) bool {
			if sr, ok := resource.(resources.StackResource); ok {
				stacks = append(stacks, *sr.ToFields())
			}
			return true
		})
	return stacks, apiErr
}
