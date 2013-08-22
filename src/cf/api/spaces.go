package api

import (
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"strings"
)

type SpaceRepository interface {
	FindSpaces(config *configuration.Configuration) (spaces []cf.Space, err error)
	FindSpaceByName(config *configuration.Configuration, name string) (space cf.Space, err error)
}

type CloudControllerSpaceRepository struct {
}

func (repo CloudControllerSpaceRepository) FindSpaces(config *configuration.Configuration) (spaces []cf.Space, err error) {
	path := fmt.Sprintf("%s/v2/organizations/%s/spaces", config.Target, config.Organization.Guid)
	request, err := NewAuthorizedRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	response := new(ApiResponse)

	err = PerformRequestForBody(request, response)

	if err != nil {
		return
	}

	for _, r := range response.Resources {
		spaces = append(spaces, cf.Space{r.Entity.Name, r.Metadata.Guid})
	}

	return
}

func (repo CloudControllerSpaceRepository) FindSpaceByName(config *configuration.Configuration, name string) (space cf.Space, err error) {
	spaces, err := repo.FindSpaces(config)
	lowerName := strings.ToLower(name)

	if err != nil {
		return
	}

	for _, s := range spaces {
		if strings.ToLower(s.Name) == lowerName {
			return s, nil
		}
	}

	err = errors.New("Space not found")
	return
}
