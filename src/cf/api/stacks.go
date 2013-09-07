package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
)

type StackRepository interface {
	FindByName(name string) (stack cf.Stack, err error)
	FindAll() (stacks []cf.Stack, err error)
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

func (repo CloudControllerStackRepository) FindByName(name string) (stack cf.Stack, err error) {
	path := fmt.Sprintf("%s/v2/stacks?q=name%s", repo.config.Target, "%3A"+name)
	request, err := NewRequest("GET", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	findResponse := new(ApiResponse)
	_, err = repo.apiClient.PerformRequestAndParseResponse(request, findResponse)
	if err != nil {
		return
	}

	if len(findResponse.Resources) == 0 {
		err = errors.New(fmt.Sprintf("Stack %s not found", name))
		return
	}

	res := findResponse.Resources[0]
	stack.Guid = res.Metadata.Guid
	stack.Name = res.Entity.Name

	return
}

func (repo CloudControllerStackRepository) FindAll() (stacks []cf.Stack, err error) {
	path := fmt.Sprintf("%s/v2/stacks", repo.config.Target)
	request, err := NewRequest("GET", path, repo.config.AccessToken, nil)
	if err != nil {
		return
	}

	listResponse := new(StackApiResponse)
	_, err = repo.apiClient.PerformRequestAndParseResponse(request, listResponse)
	if err != nil {
		return
	}

	for _, r := range listResponse.Resources {
		stacks = append(stacks, cf.Stack{Guid: r.Metadata.Guid, Name: r.Entity.Name, Description: r.Entity.Description})
	}

	return
}
