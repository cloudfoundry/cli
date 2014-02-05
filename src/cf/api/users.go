package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	neturl "net/url"
	"strings"
)

type PaginatedUserResources struct {
	NextUrl   string `json:"next_url"`
	Resources []UserResource
}

type UserResource struct {
	Resource
	Entity UserEntity
}

type UserEntity struct {
	Entity
	Admin bool
}

var orgRoleToPathMap = map[string]string{
	cf.ORG_USER:        "users",
	cf.ORG_MANAGER:     "managers",
	cf.BILLING_MANAGER: "billing_managers",
	cf.ORG_AUDITOR:     "auditors",
}

var spaceRoleToPathMap = map[string]string{
	cf.SPACE_MANAGER:   "managers",
	cf.SPACE_DEVELOPER: "developers",
	cf.SPACE_AUDITOR:   "auditors",
}

type UserRepository interface {
	FindByUsername(username string) (user cf.UserFields, apiResponse net.ApiResponse)
	ListUsersInOrgForRole(orgGuid string, role string, stop chan bool) (usersChan chan []cf.UserFields, statusChan chan net.ApiResponse)
	ListUsersInSpaceForRole(spaceGuid string, role string, stop chan bool) (usersChan chan []cf.UserFields, statusChan chan net.ApiResponse)
	Create(username, password string) (apiResponse net.ApiResponse)
	Delete(userGuid string) (apiResponse net.ApiResponse)
	SetOrgRole(userGuid, orgGuid, role string) (apiResponse net.ApiResponse)
	UnsetOrgRole(userGuid, orgGuid, role string) (apiResponse net.ApiResponse)
	SetSpaceRole(userGuid, spaceGuid, orgGuid, role string) (apiResponse net.ApiResponse)
	UnsetSpaceRole(userGuid, spaceGuid, role string) (apiResponse net.ApiResponse)
}

type CloudControllerUserRepository struct {
	config       *configuration.Configuration
	uaaGateway   net.Gateway
	ccGateway    net.Gateway
	endpointRepo EndpointRepository
}

func NewCloudControllerUserRepository(config *configuration.Configuration, uaaGateway net.Gateway, ccGateway net.Gateway, endpointRepo EndpointRepository) (repo CloudControllerUserRepository) {
	repo.config = config
	repo.uaaGateway = uaaGateway
	repo.ccGateway = ccGateway
	repo.endpointRepo = endpointRepo
	return
}

func (repo CloudControllerUserRepository) FindByUsername(username string) (user cf.UserFields, apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetUAAEndpoint()
	if apiResponse.IsNotSuccessful() {
		return
	}

	usernameFilter := neturl.QueryEscape(fmt.Sprintf(`userName Eq "%s"`, username))
	path := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, usernameFilter)

	users, apiResponse := repo.updateOrFindUsersWithUAAPath([]cf.UserFields{}, path)
	if len(users) == 0 {
		apiResponse = net.NewNotFoundApiResponse("UserFields %s not found", username)
		return
	}

	user = users[0]
	return
}

func (repo CloudControllerUserRepository) ListUsersInOrgForRole(orgGuid string, roleName string, stop chan bool) (usersChan chan []cf.UserFields, statusChan chan net.ApiResponse) {
	path := fmt.Sprintf("/v2/organizations/%s/%s", orgGuid, orgRoleToPathMap[roleName])
	return repo.listUsersForRole(path, roleName, stop)
}

func (repo CloudControllerUserRepository) ListUsersInSpaceForRole(spaceGuid string, roleName string, stop chan bool) (usersChan chan []cf.UserFields, statusChan chan net.ApiResponse) {
	path := fmt.Sprintf("/v2/spaces/%s/%s", spaceGuid, spaceRoleToPathMap[roleName])
	return repo.listUsersForRole(path, roleName, stop)
}

func (repo CloudControllerUserRepository) listUsersForRole(path string, roleName string, stop chan bool) (usersChan chan []cf.UserFields, statusChan chan net.ApiResponse) {
	usersChan = make(chan []cf.UserFields, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
	loop:
		for path != "" {
			select {
			case <-stop:
				break loop
			default:
				var (
					users       []cf.UserFields
					apiResponse net.ApiResponse
				)

				users, path, apiResponse = repo.findNextWithPath(path)
				if apiResponse.IsNotSuccessful() {
					statusChan <- apiResponse
					close(usersChan)
					close(statusChan)
					return
				}

				if len(users) > 0 {
					usersChan <- users
				}
			}
		}
		close(usersChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerUserRepository) findNextWithPath(path string) (users []cf.UserFields, nextUrl string, apiResponse net.ApiResponse) {
	paginatedResources := new(PaginatedUserResources)

	apiResponse = repo.ccGateway.GetResource(repo.config.Target+path, repo.config.AccessToken, paginatedResources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = paginatedResources.NextUrl

	if len(paginatedResources.Resources) == 0 {
		return
	}

	uaaEndpoint, apiResponse := repo.endpointRepo.GetUAAEndpoint()
	if apiResponse.IsNotSuccessful() {
		return
	}

	guidFilters := []string{}
	for _, r := range paginatedResources.Resources {
		users = append(users, cf.UserFields{Guid: r.Metadata.Guid, IsAdmin: r.Entity.Admin})
		guidFilters = append(guidFilters, fmt.Sprintf(`Id eq "%s"`, r.Metadata.Guid))
	}
	filter := strings.Join(guidFilters, " or ")
	url := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, neturl.QueryEscape(filter))

	users, apiResponse = repo.updateOrFindUsersWithUAAPath(users, url)
	return
}

func (repo CloudControllerUserRepository) updateOrFindUsersWithUAAPath(ccUsers []cf.UserFields, path string) (updatedUsers []cf.UserFields, apiResponse net.ApiResponse) {
	type uaaUserResource struct {
		Id       string
		Username string
	}
	type uaaUserResources struct {
		Resources []uaaUserResource
	}

	uaaResponse := new(uaaUserResources)
	apiResponse = repo.uaaGateway.GetResource(path, repo.config.AccessToken, uaaResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, uaaResource := range uaaResponse.Resources {
		var ccUserFields cf.UserFields

		for _, u := range ccUsers {
			if u.Guid == uaaResource.Id {
				ccUserFields = u
				break
			}
		}

		updatedUsers = append(updatedUsers, cf.UserFields{
			Guid:     uaaResource.Id,
			Username: uaaResource.Username,
			IsAdmin:  ccUserFields.IsAdmin,
		})
	}
	return
}

func (repo CloudControllerUserRepository) Create(username, password string) (apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetUAAEndpoint()
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/Users", uaaEndpoint)
	body := fmt.Sprintf(`{
  "userName": "%s",
  "emails": [{"value":"%s"}],
  "password": "%s",
  "name": {"givenName":"%s", "familyName":"%s"}
}`,
		username,
		username,
		password,
		username,
		username,
	)
	request, apiResponse := repo.uaaGateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	type uaaUserFields struct {
		Id string
	}
	createUserResponse := &uaaUserFields{}

	_, apiResponse = repo.uaaGateway.PerformRequestForJSONResponse(request, createUserResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path = fmt.Sprintf("%s/v2/users", repo.config.Target)
	body = fmt.Sprintf(`{"guid":"%s"}`, createUserResponse.Id)
	return repo.ccGateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerUserRepository) Delete(userGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/users/%s", repo.config.Target, userGuid)

	apiResponse = repo.ccGateway.DeleteResource(path, repo.config.AccessToken)
	if apiResponse.IsNotSuccessful() && apiResponse.ErrorCode != cf.USER_NOT_FOUND {
		return
	}

	uaaEndpoint, apiResponse := repo.endpointRepo.GetUAAEndpoint()
	if apiResponse.IsNotSuccessful() {
		return
	}

	path = fmt.Sprintf("%s/Users/%s", uaaEndpoint, userGuid)
	return repo.uaaGateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerUserRepository) SetOrgRole(userGuid string, orgGuid string, role string) (apiResponse net.ApiResponse) {
	apiResponse = repo.setOrUnsetOrgRole("PUT", userGuid, orgGuid, role)
	if apiResponse.IsNotSuccessful() {
		return
	}
	return repo.addOrgUserRole(userGuid, orgGuid)
}

func (repo CloudControllerUserRepository) UnsetOrgRole(userGuid, orgGuid, role string) (apiResponse net.ApiResponse) {
	return repo.setOrUnsetOrgRole("DELETE", userGuid, orgGuid, role)
}

func (repo CloudControllerUserRepository) setOrUnsetOrgRole(verb, userGuid, orgGuid, role string) (apiResponse net.ApiResponse) {
	rolePath, found := orgRoleToPathMap[role]

	if !found {
		apiResponse = net.NewApiResponseWithMessage("Invalid Role %s", role)
		return
	}

	path := fmt.Sprintf("%s/v2/organizations/%s/%s/%s", repo.config.Target, orgGuid, rolePath, userGuid)

	request, apiResponse := repo.ccGateway.NewRequest(verb, path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.ccGateway.PerformRequest(request)
	if apiResponse.IsNotSuccessful() {
		return
	}
	return
}

func (repo CloudControllerUserRepository) SetSpaceRole(userGuid, spaceGuid, orgGuid, role string) (apiResponse net.ApiResponse) {
	rolePath, apiResponse := repo.checkSpaceRole(userGuid, spaceGuid, role)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.addOrgUserRole(userGuid, orgGuid)
	if apiResponse.IsNotSuccessful() {
		return
	}

	return repo.ccGateway.UpdateResource(rolePath, repo.config.AccessToken, nil)
}

func (repo CloudControllerUserRepository) UnsetSpaceRole(userGuid, spaceGuid, role string) (apiResponse net.ApiResponse) {
	rolePath, apiResponse := repo.checkSpaceRole(userGuid, spaceGuid, role)
	if apiResponse.IsNotSuccessful() {
		return
	}
	return repo.ccGateway.DeleteResource(rolePath, repo.config.AccessToken)
}

func (repo CloudControllerUserRepository) checkSpaceRole(userGuid, spaceGuid, role string) (fullPath string, apiResponse net.ApiResponse) {
	rolePath, found := spaceRoleToPathMap[role]

	if !found {
		apiResponse = net.NewApiResponseWithMessage("Invalid Role %s", role)
	}

	fullPath = fmt.Sprintf("%s/v2/spaces/%s/%s/%s", repo.config.Target, spaceGuid, rolePath, userGuid)
	return
}

func (repo CloudControllerUserRepository) addOrgUserRole(userGuid, orgGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s/users/%s", repo.config.Target, orgGuid, userGuid)
	return repo.ccGateway.UpdateResource(path, repo.config.AccessToken, nil)
}
