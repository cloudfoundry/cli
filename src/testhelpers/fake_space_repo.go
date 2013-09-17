package testhelpers

import (
	"cf"
	"cf/net"
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

	DeletedSpace cf.Space
}

func (repo FakeSpaceRepository) GetCurrentSpace() (space cf.Space) {
	return repo.CurrentSpace
}

func (repo FakeSpaceRepository) FindAll() (spaces []cf.Space, apiErr *net.ApiError) {
	return repo.Spaces, nil
}

func (repo *FakeSpaceRepository) FindByName(name string) (space cf.Space, apiErr *net.ApiError) {
	repo.SpaceName = name
	if repo.SpaceByNameErr {
		apiErr = net.NewApiErrorWithMessage("Error finding space by name.")
	}
	return repo.SpaceByName, apiErr
}

func (repo *FakeSpaceRepository) GetSummary() (space cf.Space, apiErr *net.ApiError) {
	space = repo.SummarySpace
	return
}

func (repo *FakeSpaceRepository) Create(name string) (apiErr *net.ApiError) {
	repo.CreateSpaceName = name
	return
}

func (repo *FakeSpaceRepository) Rename(space cf.Space, newName string) (apiErr *net.ApiError) {
	repo.RenameSpace = space
	repo.RenameNewName = newName
	return
}

func (repo *FakeSpaceRepository) Delete(space cf.Space) (apiErr *net.ApiError) {
	repo.DeletedSpace = space
	return
}
