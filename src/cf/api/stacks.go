package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
)

type PaginatedStackResources struct {
	Resources []StackResource
}

type StackResource struct {
	Resource
	Entity StackEntity
}

func (resource StackResource) ToFields() (fields models.Stack) {
	fields.Guid = resource.Metadata.Guid
	fields.Name = resource.Entity.Name
	fields.Description = resource.Entity.Description
	return
}

type StackEntity struct {
	Name        string
	Description string
}

type StackRepository interface {
	FindByName(name string) (stack models.Stack, apiResponse errors.Error)
	FindAll() (stacks []models.Stack, apiResponse errors.Error)
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

func (repo CloudControllerStackRepository) FindByName(name string) (stack models.Stack, apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/stacks?q=%s", repo.config.ApiEndpoint(), url.QueryEscape("name:"+name))
	stacks, apiResponse := repo.findAllWithPath(path)
	if apiResponse != nil {
		return
	}

	if len(stacks) == 0 {
		apiResponse = errors.NewErrorWithMessage("Stack '%s' not found", name)
		return
	}

	stack = stacks[0]
	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []models.Stack, apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.ApiEndpoint())
	return repo.findAllWithPath(path)
}

func (repo CloudControllerStackRepository) findAllWithPath(path string) (stacks []models.Stack, apiResponse errors.Error) {
	resources := new(PaginatedStackResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken(), resources)
	if apiResponse != nil {
		return
	}

	for _, r := range resources.Resources {
		stacks = append(stacks, r.ToFields())
	}
	return
}
