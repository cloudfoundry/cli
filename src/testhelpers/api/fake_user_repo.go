package api

import (
	"cf"
	"cf/errors"
	"cf/models"
)

type FakeUserRepository struct {
	FindByUsernameUsername   string
	FindByUsernameUserFields models.UserFields
	FindByUsernameNotFound   bool

	ListUsersOrganizationGuid string
	ListUsersSpaceGuid        string
	ListUsersByRole           map[string][]models.UserFields

	CreateUserUsername string
	CreateUserPassword string
	CreateUserExists   bool

	DeleteUserGuid string

	SetOrgRoleUserGuid         string
	SetOrgRoleOrganizationGuid string
	SetOrgRoleRole             string

	UnsetOrgRoleUserGuid         string
	UnsetOrgRoleOrganizationGuid string
	UnsetOrgRoleRole             string

	SetSpaceRoleUserGuid  string
	SetSpaceRoleOrgGuid   string
	SetSpaceRoleSpaceGuid string
	SetSpaceRoleRole      string

	UnsetSpaceRoleUserGuid  string
	UnsetSpaceRoleSpaceGuid string
	UnsetSpaceRoleRole      string
}

func (repo *FakeUserRepository) FindByUsername(username string) (user models.UserFields, apiErr errors.Error) {
	repo.FindByUsernameUsername = username
	user = repo.FindByUsernameUserFields

	if repo.FindByUsernameNotFound {
		apiErr = errors.NewModelNotFoundError("User", "")
	}

	return
}

func (repo *FakeUserRepository) ListUsersInOrgForRole(orgGuid string, roleName string) ([]models.UserFields, errors.Error) {
	repo.ListUsersOrganizationGuid = orgGuid
	return repo.ListUsersByRole[roleName], nil
}

func (repo *FakeUserRepository) ListUsersInSpaceForRole(spaceGuid string, roleName string) ([]models.UserFields, errors.Error) {
	repo.ListUsersSpaceGuid = spaceGuid
	return repo.ListUsersByRole[roleName], nil
}

func (repo *FakeUserRepository) Create(username, password string) (apiErr errors.Error) {
	repo.CreateUserUsername = username
	repo.CreateUserPassword = password

	if repo.CreateUserExists {
		apiErr = errors.NewError("User already exists", cf.USER_EXISTS)
	}

	return
}

func (repo *FakeUserRepository) Delete(userGuid string) (apiErr errors.Error) {
	repo.DeleteUserGuid = userGuid
	return
}

func (repo *FakeUserRepository) SetOrgRole(userGuid, orgGuid, role string) (apiErr errors.Error) {
	repo.SetOrgRoleUserGuid = userGuid
	repo.SetOrgRoleOrganizationGuid = orgGuid
	repo.SetOrgRoleRole = role
	return
}

func (repo *FakeUserRepository) UnsetOrgRole(userGuid, orgGuid, role string) (apiErr errors.Error) {
	repo.UnsetOrgRoleUserGuid = userGuid
	repo.UnsetOrgRoleOrganizationGuid = orgGuid
	repo.UnsetOrgRoleRole = role
	return
}

func (repo *FakeUserRepository) SetSpaceRole(userGuid, spaceGuid, orgGuid, role string) (apiErr errors.Error) {
	repo.SetSpaceRoleUserGuid = userGuid
	repo.SetSpaceRoleOrgGuid = orgGuid
	repo.SetSpaceRoleSpaceGuid = spaceGuid
	repo.SetSpaceRoleRole = role
	return
}

func (repo *FakeUserRepository) UnsetSpaceRole(userGuid, spaceGuid, role string) (apiErr errors.Error) {
	repo.UnsetSpaceRoleUserGuid = userGuid
	repo.UnsetSpaceRoleSpaceGuid = spaceGuid
	repo.UnsetSpaceRoleRole = role
	return
}
