package api

import (
	"cf"
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	neturl "net/url"
	"strings"
)

type UserResource struct {
	Resource
	Entity UserEntity
}

type UAAUserResources struct {
	Resources []struct {
		Id       string
		Username string
	}
}

func (resource UserResource) ToFields() models.UserFields {
	return models.UserFields{
		Guid:    resource.Metadata.Guid,
		IsAdmin: resource.Entity.Admin,
	}
}

type UserEntity struct {
	Entity
	Admin bool
}

var orgRoleToPathMap = map[string]string{
	models.ORG_USER:        "users",
	models.ORG_MANAGER:     "managers",
	models.BILLING_MANAGER: "billing_managers",
	models.ORG_AUDITOR:     "auditors",
}

var spaceRoleToPathMap = map[string]string{
	models.SPACE_MANAGER:   "managers",
	models.SPACE_DEVELOPER: "developers",
	models.SPACE_AUDITOR:   "auditors",
}

type UserRepository interface {
	FindByUsername(username string) (user models.UserFields, apiResponse net.ApiResponse)
	ListUsersInOrgForRole(orgGuid string, role string) ([]models.UserFields, net.ApiResponse)
	ListUsersInSpaceForRole(spaceGuid string, role string) ([]models.UserFields, net.ApiResponse)
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

func (repo CloudControllerUserRepository) FindByUsername(username string) (user models.UserFields, apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetUAAEndpoint()
	if apiResponse.IsNotSuccessful() {
		return
	}

	usernameFilter := neturl.QueryEscape(fmt.Sprintf(`userName Eq "%s"`, username))
	path := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, usernameFilter)

	users, apiResponse := repo.updateOrFindUsersWithUAAPath([]models.UserFields{}, path)
	if len(users) == 0 {
		apiResponse = net.NewNotFoundApiResponse("UserFields %s not found", username)
		return
	}

	user = users[0]
	return
}

func (repo CloudControllerUserRepository) ListUsersInOrgForRole(orgGuid string, roleName string) (users []models.UserFields, apiResponse net.ApiResponse) {
	return repo.listUsersWithPath(fmt.Sprintf("/v2/organizations/%s/%s", orgGuid, orgRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) ListUsersInSpaceForRole(spaceGuid string, roleName string) (users []models.UserFields, apiResponse net.ApiResponse) {
	return repo.listUsersWithPath(fmt.Sprintf("/v2/spaces/%s/%s", spaceGuid, spaceRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) listUsersWithPath(path string) (users []models.UserFields, apiResponse net.ApiResponse) {
	guidFilters := []string{}

	apiResponse = repo.ccGateway.ListPaginatedResources(
		repo.config.Target,
		repo.config.AccessToken,
		path,
		UserResource{},
		func(resource interface{}) bool {
			user := resource.(UserResource).ToFields()
			users = append(users, user)
			guidFilters = append(guidFilters, fmt.Sprintf(`Id eq "%s"`, user.Guid))
			return true
		})
	if apiResponse.IsNotSuccessful() {
		return
	}

	uaaEndpoint, apiResponse := repo.endpointRepo.GetUAAEndpoint()
	if apiResponse.IsNotSuccessful() {
		return
	}

	filter := strings.Join(guidFilters, " or ")
	usersURL := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, neturl.QueryEscape(filter))
	users, apiResponse = repo.updateOrFindUsersWithUAAPath(users, usersURL)
	return
}

func (repo CloudControllerUserRepository) updateOrFindUsersWithUAAPath(ccUsers []models.UserFields, path string) (updatedUsers []models.UserFields, apiResponse net.ApiResponse) {
	uaaResponse := new(UAAUserResources)
	apiResponse = repo.uaaGateway.GetResource(path, repo.config.AccessToken, uaaResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, uaaResource := range uaaResponse.Resources {
		var ccUserFields models.UserFields

		for _, u := range ccUsers {
			if u.Guid == uaaResource.Id {
				ccUserFields = u
				break
			}
		}

		updatedUsers = append(updatedUsers, models.UserFields{
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
