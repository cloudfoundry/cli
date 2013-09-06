package testhelpers

import (
	"errors"
	"cf"
)

type FakeSpaceRepository struct {
	CurrentSpace cf.Space

	Spaces []cf.Space

	SpaceName string
	SpaceByName cf.Space
	SpaceByNameErr bool

	SummarySpace cf.Space
}

func (repo FakeSpaceRepository) GetCurrentSpace() (space cf.Space) {
	return repo.CurrentSpace
}

func (repo FakeSpaceRepository) FindAll() (spaces []cf.Space, err error) {
	return repo.Spaces, nil
}

func (repo *FakeSpaceRepository) FindByName(name string) (space cf.Space, err error) {
	repo.SpaceName = name
	if repo.SpaceByNameErr {
		err = errors.New("Error finding space by name.")
	}
	return repo.SpaceByName, err
}

func (repo *FakeSpaceRepository) GetSummary() (space cf.Space, err error) {
	space = repo.SummarySpace
	return
}
