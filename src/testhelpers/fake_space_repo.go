package testhelpers

import (
	"cf"
	"cf/api"
)

type FakeSpaceRepository struct {
	CurrentSpace cf.Space

	Spaces []cf.Space

	SpaceName string
	SpaceByName cf.Space
	SpaceByNameErr bool

	SummarySpace cf.Space

	CreateSpaceName string

	RenameSpace cf.Space
	RenameNewName string
}

func (repo FakeSpaceRepository) GetCurrentSpace() (space cf.Space) {
	return repo.CurrentSpace
}

func (repo FakeSpaceRepository) FindAll() (spaces []cf.Space, apiErr *api.ApiError) {
	return repo.Spaces, nil
}

func (repo *FakeSpaceRepository) FindByName(name string) (space cf.Space, apiErr *api.ApiError) {
	repo.SpaceName = name
	if repo.SpaceByNameErr {
		apiErr = api.NewApiErrorWithMessage("Error finding space by name.")
	}
	return repo.SpaceByName, apiErr
}

func (repo *FakeSpaceRepository) GetSummary() (space cf.Space, apiErr *api.ApiError) {
	space = repo.SummarySpace
	return
}

func (repo *FakeSpaceRepository) Create(name string) (apiErr *api.ApiError) {
	repo.CreateSpaceName = name
	return
}

func (repo *FakeSpaceRepository) Rename(space cf.Space, newName string) (apiErr *api.ApiError) {
	repo.RenameSpace = space
	repo.RenameNewName = newName
	return
}
