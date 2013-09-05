package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
)

type StackRepository interface {
	FindByName(config *configuration.Configuration, name string) (stack cf.Stack, err error)
}

type CloudControllerStackRepository struct {
}

func (repo CloudControllerStackRepository) FindByName(config *configuration.Configuration, name string) (stack cf.Stack, err error) {
	path := fmt.Sprintf("%s/v2/stacks?q=name%s", config.Target, "%3A"+name)
	request, err := NewRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	findResponse := new(ApiResponse)
	_, err = PerformRequestAndParseResponse(request, findResponse)
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
