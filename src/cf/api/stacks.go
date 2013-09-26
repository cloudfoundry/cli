package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
)

type StackRepository interface {
	FindByName(name string) (stack cf.Stack, apiErr *net.ApiError)
	FindAll() (stacks []cf.Stack, apiErr *net.ApiError)
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

func (repo CloudControllerStackRepository) FindByName(name string) (stack cf.Stack, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/stacks?q=name%s", repo.config.Target, "%3A"+name)
	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	findResponse := new(ApiResponse)
	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, findResponse)
	if apiErr != nil {
		return
	}

	if len(findResponse.Resources) == 0 {
		apiErr = net.NewApiErrorWithMessage("Stack %s not found", name)
		return
	}

	res := findResponse.Resources[0]
	stack.Guid = res.Metadata.Guid
	stack.Name = res.Entity.Name

	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []cf.Stack, apiErr *net.ApiError) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.Target)
	request, apiErr := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	listResponse := new(StackApiResponse)
	_, apiErr = repo.gateway.PerformRequestForJSONResponse(request, listResponse)
	if apiErr != nil {
		return
	}

	for _, r := range listResponse.Resources {
		stacks = append(stacks, cf.Stack{Guid: r.Metadata.Guid, Name: r.Entity.Name, Description: r.Entity.Description})
	}

	return
}
