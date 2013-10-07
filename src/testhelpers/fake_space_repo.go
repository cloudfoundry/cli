package testhelpers

import (
	"cf"
	"cf/net"
	"cf/api"
)

type FakeSpaceRepository struct {
	CurrentSpace cf.Space

	Spaces []cf.Space

	FindByNameName string
	FindByNameSpace cf.Space
	FindByNameErr bool
	FindByNameNotFound bool

	SummarySpace cf.Space

	CreateSpaceName string
	CreateSpaceExists bool

	RenameSpace cf.Space
	RenameNewName string

	DeletedSpace cf.Space
}

func (repo FakeSpaceRepository) GetCurrentSpace() (space cf.Space) {
	return repo.CurrentSpace
}

func (repo FakeSpaceRepository) FindAll() (spaces []cf.Space, apiResponse net.ApiResponse) {
	spaces = repo.Spaces
	return
}

func (repo *FakeSpaceRepository) FindByName(name string) (space cf.Space, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	space = repo.FindByNameSpace

	if repo.FindByNameErr {
		apiResponse = net.NewApiStatusWithMessage("Error finding space by name.")
	}

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiStatus("Space", name)
	}

	return
}

func (repo *FakeSpaceRepository) GetSummary() (space cf.Space, apiResponse net.ApiResponse) {
	space = repo.SummarySpace
	return
}

func (repo *FakeSpaceRepository) Create(name string) (apiResponse net.ApiResponse) {
	if repo.CreateSpaceExists {
		apiResponse = net.NewApiStatus("Space already exists", api.SPACE_EXISTS, 400)
		return
	}
	repo.CreateSpaceName = name
	return
}

func (repo *FakeSpaceRepository) Rename(space cf.Space, newName string) (apiResponse net.ApiResponse) {
	repo.RenameSpace = space
	repo.RenameNewName = newName
	return
}

func (repo *FakeSpaceRepository) Delete(space cf.Space) (apiResponse net.ApiResponse) {
	repo.DeletedSpace = space
	return
}
