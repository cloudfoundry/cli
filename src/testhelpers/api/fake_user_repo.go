package api

import (
	"cf/net"
	"cf"
)

type FakeUserRepository struct {
	FindByUsernameUsername string
	FindByUsernameUserFields cf.UserFields
	FindByUsernameNotFound bool

	ListUsersOrganizationGuid string
	ListUsersSpaceGuid string
	ListUsersByRole map[string][]cf.UserFields

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
	SetSpaceRoleOrgGuid string
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

func (repo *FakeUserRepository) ListUsersInOrgForRole(orgGuid string, roleName string, stop chan bool) (usersChan chan []cf.UserFields, statusChan chan net.ApiResponse) {
	repo.ListUsersOrganizationGuid = orgGuid
	return repo.listUsersForRole(roleName,stop)
}

func (repo *FakeUserRepository) ListUsersInSpaceForRole(spaceGuid string, roleName string, stop chan bool) (usersChan chan []cf.UserFields, statusChan chan net.ApiResponse){
	repo.ListUsersSpaceGuid = spaceGuid
	return repo.listUsersForRole(roleName,stop)
}

func (repo *FakeUserRepository) listUsersForRole(roleName string, stop chan bool) (usersChan chan []cf.UserFields, statusChan chan net.ApiResponse){
	usersChan = make(chan []cf.UserFields, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		usersCount := len(repo.ListUsersByRole[roleName])
		for i:= 0; i < usersCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if usersCount - i > 1 {
					usersChan <- repo.ListUsersByRole[roleName][i:i+2]
				} else {
					usersChan <- repo.ListUsersByRole[roleName][i:]
				}
			}
		}

		close(usersChan)
		close(statusChan)

		cf.WaitForClose(stop)
	}()

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
	repo.SetSpaceRoleOrgGuid = orgGuid
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
