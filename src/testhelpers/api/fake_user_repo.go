package api

import (
	"cf/net"
	"cf"
)

type FakeUserRepository struct {
	FindByUsernameUsername string
	FindByUsernameUser cf.User
	FindByUsernameNotFound bool

	FindAllInOrgByRoleOrganization cf.Organization
	FindAllInOrgByRoleUsersByRole map[string][]cf.User

	CreateUserUser cf.User
	CreateUserExists bool

	DeleteUser cf.User

	SetOrgRoleUser cf.User
	SetOrgRoleOrganization cf.Organization
	SetOrgRoleRole string

	UnsetOrgRoleUser cf.User
	UnsetOrgRoleOrganization cf.Organization
	UnsetOrgRoleRole string

	SetSpaceRoleUser cf.User
	SetSpaceRoleSpace cf.Space
	SetSpaceRoleRole string

	UnsetSpaceRoleUser cf.User
	UnsetSpaceRoleSpace cf.Space
	UnsetSpaceRoleRole string
}

func (repo *FakeUserRepository) FindByUsername(username string) (user cf.User, apiResponse net.ApiResponse) {
	repo.FindByUsernameUsername = username
	user = repo.FindByUsernameUser

	if repo.FindByUsernameNotFound {
		apiResponse = net.NewNotFoundApiResponse("User not found")
	}

	return
}

func (repo *FakeUserRepository) FindAllInOrgByRole(org cf.Organization) (usersByRole map[string][]cf.User, apiResponse net.ApiResponse) {
	repo.FindAllInOrgByRoleOrganization = org
	usersByRole = repo.FindAllInOrgByRoleUsersByRole
	return
}

func (repo *FakeUserRepository) Create(user cf.User) (apiResponse net.ApiResponse) {
	repo.CreateUserUser = user

	if repo.CreateUserExists {
		apiResponse = net.NewApiResponse("User already exists", cf.USER_EXISTS, 400)
	}

	return
}

func (repo *FakeUserRepository) Delete(user cf.User) (apiResponse net.ApiResponse) {
	repo.DeleteUser = user
	return
}

func (repo *FakeUserRepository) SetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse) {
	repo.SetOrgRoleUser = user
	repo.SetOrgRoleOrganization = org
	repo.SetOrgRoleRole = role
	return
}

func (repo *FakeUserRepository) UnsetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse) {
	repo.UnsetOrgRoleUser = user
	repo.UnsetOrgRoleOrganization = org
	repo.UnsetOrgRoleRole = role
	return
}

func (repo *FakeUserRepository) SetSpaceRole(user cf.User, space cf.Space, role string) (apiResponse net.ApiResponse) {
	repo.SetSpaceRoleUser = user
	repo.SetSpaceRoleSpace = space
	repo.SetSpaceRoleRole = role
	return
}

func (repo *FakeUserRepository) UnsetSpaceRole(user cf.User, space cf.Space, role string) (apiResponse net.ApiResponse) {
	repo.UnsetSpaceRoleUser = user
	repo.UnsetSpaceRoleSpace = space
	repo.UnsetSpaceRoleRole = role
	return
}
