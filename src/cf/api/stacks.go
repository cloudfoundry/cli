package api

import (
	"cf/api/resources"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
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
	path := fmt.Sprintf("%s/v2/stacks?q=%s", repo.config.ApiEndpoint(), url.QueryEscape("name:"+name))
	stacks, apiErr := repo.findAllWithPath(path)
	if apiErr != nil {
		return
	}

	if len(stacks) == 0 {
		apiErr = errors.NewWithFmt("Stack '%s' not found", name)
		return
	}

	stack = stacks[0]
	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []models.Stack, apiErr error) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.ApiEndpoint())
	return repo.findAllWithPath(path)
}

func (repo CloudControllerStackRepository) findAllWithPath(path string) (stacks []models.Stack, apiErr error) {
	responseJSON := new(resources.PaginatedStackResources)
	apiErr = repo.gateway.GetResource(path, responseJSON)
	if apiErr != nil {
		return
	}

	for _, r := range responseJSON.Resources {
		stacks = append(stacks, *r.ToFields())
	}
	return
}
