package testhelpers

import (
	"cf"
	"cf/net"
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

func (repo FakeSpaceRepository) FindAll() (spaces []cf.Space, apiStatus net.ApiStatus) {
	spaces = repo.Spaces
	return
}

func (repo *FakeSpaceRepository) FindByName(name string) (space cf.Space, apiStatus net.ApiStatus) {
	repo.FindByNameName = name
	space = repo.FindByNameSpace

	if repo.FindByNameErr {
		apiStatus = net.NewApiStatusWithMessage("Error finding space by name.")
	}

	return
}

func (repo *FakeSpaceRepository) GetSummary() (space cf.Space, apiStatus net.ApiStatus) {
	space = repo.SummarySpace
	return
}

func (repo *FakeSpaceRepository) Create(name string) (apiStatus net.ApiStatus) {
	if repo.CreateSpaceExists {
		apiStatus = net.NewApiStatus("Space already exists", net.SPACE_EXISTS, 400)
		return
	}
	repo.CreateSpaceName = name
	return
}

func (repo *FakeSpaceRepository) Rename(space cf.Space, newName string) (apiStatus net.ApiStatus) {
	repo.RenameSpace = space
	repo.RenameNewName = newName
	return
}

func (repo *FakeSpaceRepository) Delete(space cf.Space) (apiStatus net.ApiStatus) {
	repo.DeletedSpace = space
	return
}
