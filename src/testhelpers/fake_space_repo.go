package testhelpers

import (
	"errors"
	"cf"
	"cf/configuration"
)

type FakeSpaceRepository struct {
	Spaces []cf.Space

	SpaceName string
	SpaceByName cf.Space
	SpaceByNameErr bool

	SummarySpace cf.Space
}

func (repo FakeSpaceRepository) FindAll(config *configuration.Configuration) (spaces []cf.Space, err error) {
	return repo.Spaces, nil
}

func (repo *FakeSpaceRepository) FindByName(config *configuration.Configuration, name string) (space cf.Space, err error) {
	repo.SpaceName = name
	if repo.SpaceByNameErr {
		err = errors.New("Error finding space by name.")
	}
	return repo.SpaceByName, err
}

func (repo *FakeSpaceRepository) GetSummary(config *configuration.Configuration) (space cf.Space, err error) {
	space = repo.SummarySpace
	return
}
