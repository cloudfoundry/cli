package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
)

type StackRepository interface {
	FindByName(name string) (stack cf.Stack, apiResponse net.ApiResponse)
	FindAll() (stacks []cf.Stack, apiResponse net.ApiResponse)
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

func (repo CloudControllerStackRepository) FindByName(name string) (stack cf.Stack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/stacks?q=name%s", repo.config.Target, "%3A"+name)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedResources)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(resources.Resources) == 0 {
		apiResponse = net.NewApiResponseWithMessage("Stack %s not found", name)
		return
	}

	res := resources.Resources[0]
	stack.Guid = res.Metadata.Guid
	stack.Name = res.Entity.Name

	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []cf.Stack, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.Target)
	request, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedStackResources)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(request, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, r := range resources.Resources {
		stacks = append(stacks, cf.Stack{Guid: r.Metadata.Guid, Name: r.Entity.Name, Description: r.Entity.Description})
	}

	return
}
