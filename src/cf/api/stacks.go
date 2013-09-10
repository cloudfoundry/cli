package api

import (
	"cf"
	"cf/configuration"
	"fmt"
)

type StackRepository interface {
	FindByName(name string) (stack cf.Stack, apiErr *ApiError)
	FindAll() (stacks []cf.Stack, apiErr *ApiError)
}

type CloudControllerStackRepository struct {
	config    *configuration.Configuration
	apiClient ApiClient
}

func NewCloudControllerStackRepository(config *configuration.Configuration, apiClient ApiClient) (repo CloudControllerStackRepository) {
	repo.config = config
	repo.apiClient = apiClient
	return
}

func (repo CloudControllerStackRepository) FindByName(name string) (stack cf.Stack, apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/stacks?q=name%s", repo.config.Target, "%3A"+name)
	request, apiErr := NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	findResponse := new(ApiResponse)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, findResponse)
	if apiErr != nil {
		return
	}

	if len(findResponse.Resources) == 0 {
		apiErr = NewApiErrorWithMessage("Stack %s not found", name)
		return
	}

	res := findResponse.Resources[0]
	stack.Guid = res.Metadata.Guid
	stack.Name = res.Entity.Name

	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []cf.Stack, apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.Target)
	request, apiErr := NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	listResponse := new(StackApiResponse)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, listResponse)
	if apiErr != nil {
		return
	}

	for _, r := range listResponse.Resources {
		stacks = append(stacks, cf.Stack{Guid: r.Metadata.Guid, Name: r.Entity.Name, Description: r.Entity.Description})
	}

	return
}
