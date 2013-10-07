package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
)

type StackRepository interface {
	FindByName(name string) (stack cf.Stack, apiStatus net.ApiStatus)
	FindAll() (stacks []cf.Stack, apiStatus net.ApiStatus)
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

func (repo CloudControllerStackRepository) FindByName(name string) (stack cf.Stack, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/stacks?q=name%s", repo.config.Target, "%3A"+name)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsNotSuccessful() {
		return
	}

	findResponse := new(ApiResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, findResponse)
	if apiStatus.IsNotSuccessful() {
		return
	}

	if len(findResponse.Resources) == 0 {
		apiStatus = net.NewApiStatusWithMessage("Stack %s not found", name)
		return
	}

	res := findResponse.Resources[0]
	stack.Guid = res.Metadata.Guid
	stack.Name = res.Entity.Name

	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []cf.Stack, apiStatus net.ApiStatus) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.Target)
	request, apiStatus := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiStatus.IsNotSuccessful() {
		return
	}

	listResponse := new(StackApiResponse)
	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, listResponse)
	if apiStatus.IsNotSuccessful() {
		return
	}

	for _, r := range listResponse.Resources {
		stacks = append(stacks, cf.Stack{Guid: r.Metadata.Guid, Name: r.Entity.Name, Description: r.Entity.Description})
	}

	return
}
