package api

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"net/url"
)

type StackRepository interface {
	FindByName(name string) (stack models.Stack, apiErr error)
	FindAll() (stacks []models.Stack, apiErr error)
}

type CloudControllerStackRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerStackRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerStackRepository) {
	repo.config = config
	repo.gateway = gateway
	return
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
