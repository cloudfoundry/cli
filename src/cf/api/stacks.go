package api

import (
	"cf/configuration"
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
	FindByName(name string) (stack models.Stack, apiResponse net.ApiResponse)
	FindAll() (stacks []models.Stack, apiResponse net.ApiResponse)
}

type CloudControllerStackRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerStackRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerStackRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerStackRepository) FindByName(name string) (stack models.Stack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/stacks?q=%s", repo.config.Target, url.QueryEscape("name:"+name))
	stacks, apiResponse := repo.findAllWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(stacks) == 0 {
		apiResponse = net.NewApiResponseWithMessage("Stack '%s' not found", name)
		return
	}

	stack = stacks[0]
	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []models.Stack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.Target)
	return repo.findAllWithPath(path)
}

func (repo CloudControllerStackRepository) findAllWithPath(path string) (stacks []models.Stack, apiResponse net.ApiResponse) {
	resources := new(PaginatedStackResources)
	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range resources.Resources {
		stacks = append(stacks, r.ToFields())
	}
	return
}
