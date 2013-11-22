package api

import (
	"cf/net"
	"cf"
)

type FakeUserRepository struct {
	FindByUsernameUsername string
	FindByUsernameUserFields cf.UserFields
	FindByUsernameNotFound bool

	FindAllInOrgByRoleOrganizationGuid string
	FindAllInOrgByRoleUsersByRole map[string][]cf.UserFields

	FindAllInSpaceByRoleSpaceGuid string
	FindAllInSpaceByRoleUsersByRole map[string][]cf.UserFields

	CreateUserUsername string
	CreateUserPassword string
	CreateUserExists bool

	DeleteUserGuid string

	SetOrgRoleUserGuid string
	SetOrgRoleOrganizationGuid string
	SetOrgRoleRole string

	UnsetOrgRoleUserGuid string
	UnsetOrgRoleOrganizationGuid string
	UnsetOrgRoleRole string

	SetSpaceRoleUserGuid string
	SetSpaceRoleSpaceGuid string
	SetSpaceRoleRole string

	UnsetSpaceRoleUserGuid string
	UnsetSpaceRoleSpaceGuid string
	UnsetSpaceRoleRole string
}

func (repo *FakeUserRepository) FindByUsername(username string) (user cf.UserFields, apiResponse net.ApiResponse) {
	repo.FindByUsernameUsername = username
	user = repo.FindByUsernameUserFields

	if repo.FindByUsernameNotFound {
		apiResponse = net.NewNotFoundApiResponse("UserFields not found")
	}

	return
}

func (repo *FakeUserRepository) FindAllInOrgByRole(orgGuid string) (usersByRole map[string][]cf.UserFields, apiResponse net.ApiResponse) {
	repo.FindAllInOrgByRoleOrganizationGuid = orgGuid
	usersByRole = repo.FindAllInOrgByRoleUsersByRole
	return
}

func (repo *FakeUserRepository) FindAllInSpaceByRole(spaceGuid string) (usersByRole map[string][]cf.UserFields, apiResponse net.ApiResponse) {
	repo.FindAllInSpaceByRoleSpaceGuid = spaceGuid
	usersByRole = repo.FindAllInSpaceByRoleUsersByRole
	return
}

func (repo *FakeUserRepository) Create(username, password string) (apiResponse net.ApiResponse) {
	repo.CreateUserUsername = username
	repo.CreateUserPassword = password

	if repo.CreateUserExists {
		apiResponse = net.NewApiResponse("UserFields already exists", cf.USER_EXISTS, 400)
	}

	return
}

func (repo *FakeUserRepository) Delete(userGuid string) (apiResponse net.ApiResponse) {
	repo.DeleteUserGuid = userGuid
	return
}

func (repo *FakeUserRepository) SetOrgRole(userGuid, orgGuid, role string) (apiResponse net.ApiResponse) {
	repo.SetOrgRoleUserGuid = userGuid
	repo.SetOrgRoleOrganizationGuid = orgGuid
	repo.SetOrgRoleRole = role
	return
}

func (repo *FakeUserRepository) UnsetOrgRole(userGuid, orgGuid, role string) (apiResponse net.ApiResponse) {
	repo.UnsetOrgRoleUserGuid = userGuid
	repo.UnsetOrgRoleOrganizationGuid = orgGuid
	repo.UnsetOrgRoleRole = role
	return
}

func (repo *FakeUserRepository) SetSpaceRole(userGuid, spaceGuid, orgGuid, role string) (apiResponse net.ApiResponse) {
	repo.SetSpaceRoleUserGuid = userGuid
	repo.SetSpaceRoleSpaceGuid = spaceGuid
	repo.SetSpaceRoleRole = role
	return
}

func (repo *FakeUserRepository) UnsetSpaceRole(userGuid, spaceGuid, role string) (apiResponse net.ApiResponse) {
	repo.UnsetSpaceRoleUserGuid = userGuid
	repo.UnsetSpaceRoleSpaceGuid = spaceGuid
	repo.UnsetSpaceRoleRole = role
	return
}
