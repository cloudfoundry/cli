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
	"OrgManager":     "managers",
	"BillingManager": "billing_managers",
	"OrgAuditor":     "auditors",
}

var spaceRoleToPathMap = map[string]string{
	"SpaceManager":   "managers",
	"SpaceDeveloper": "developers",
	"SpaceAuditor":   "auditors",
}

var orgPathToDisplayNameMap = map[string]string{
	"managers":         "ORG MANAGER",
	"billing_managers": "BILLING MANAGER",
	"auditors":         "ORG AUDITOR",
}

var spacePathToDisplayNameMap = map[string]string{
	"managers":   "SPACE MANAGER",
	"developers": "SPACE DEVELOPER",
	"auditors":   "SPACE AUDITOR",
}

type UserRepository interface {
	FindByUsername(username string) (user cf.User, apiResponse net.ApiResponse)
	FindAllInOrgByRole(org cf.Organization) (usersByRole map[string][]cf.User, apiResponse net.ApiResponse)
	FindAllInSpaceByRole(space cf.Space) (usersByRole map[string][]cf.User, apiResponse net.ApiResponse)
	Create(user cf.User) (apiResponse net.ApiResponse)
	Delete(user cf.User) (apiResponse net.ApiResponse)
	SetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse)
	UnsetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse)
	SetSpaceRole(user cf.User, org cf.Space, role string) (apiResponse net.ApiResponse)
	UnsetSpaceRole(user cf.User, org cf.Space, role string) (apiResponse net.ApiResponse)
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

func (repo CloudControllerUserRepository) FindByUsername(username string) (user cf.User, apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return
	}

	usernameFilter := neturl.QueryEscape(fmt.Sprintf(`userName Eq "%s"`, username))
	path := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, usernameFilter)

	users, apiResponse := repo.updateOrFindUsersWithUAAPath([]cf.User{}, path)
	if len(users) == 0 {
		apiResponse = net.NewNotFoundApiResponse("User %s not found", username)
		return
	}

	user = users[0]
	return
}

func (repo CloudControllerUserRepository) FindAllInOrgByRole(org cf.Organization) (usersByRole map[string][]cf.User, apiResponse net.ApiResponse) {
	usersByRole = make(map[string][]cf.User)

	for rolePath, displayName := range orgPathToDisplayNameMap {
		var users []cf.User

		path := fmt.Sprintf("/v2/organizations/%s/%s", org.Guid, rolePath)
		users, apiResponse = repo.findAllWithPath(path)
		if apiResponse.IsNotSuccessful() {
			return
		}
		usersByRole[displayName] = users
	}
	return
}

func (repo CloudControllerUserRepository) FindAllInSpaceByRole(space cf.Space) (usersByRole map[string][]cf.User, apiResponse net.ApiResponse) {
	usersByRole = make(map[string][]cf.User)

	for rolePath, displayName := range spacePathToDisplayNameMap {
		var users []cf.User

		path := fmt.Sprintf("/v2/spaces/%s/%s", space.Guid, rolePath)
		users, apiResponse = repo.findAllWithPath(path)
		if apiResponse.IsNotSuccessful() {
			return
		}
		usersByRole[displayName] = users
	}
	return
}

func (repo CloudControllerUserRepository) findAllWithPath(path string) (users []cf.User, apiResponse net.ApiResponse) {
	allUserResources := []UserResource{}

	for path != "" {
		paginatedResources := new(PaginatedUserResources)
		url := fmt.Sprintf("%s%s", repo.config.Target, path)

		apiResponse = repo.ccGateway.GetResource(url, repo.config.AccessToken, paginatedResources)
		if apiResponse.IsNotSuccessful() {
			return
		}

		allUserResources = append(allUserResources, paginatedResources.Resources...)
		path = paginatedResources.NextUrl
	}

	if len(allUserResources) == 0 {
		return
	}

	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return
	}

	guidFilters := []string{}
	for _, r := range allUserResources {
		users = append(users, cf.User{Guid: r.Metadata.Guid, IsAdmin: r.Entity.Admin})
		guidFilters = append(guidFilters, fmt.Sprintf(`Id eq "%s"`, r.Metadata.Guid))
	}
	filter := strings.Join(guidFilters, " or ")
	url := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, neturl.QueryEscape(filter))

	users, apiResponse = repo.updateOrFindUsersWithUAAPath(users, url)
	return
}

func (repo CloudControllerUserRepository) updateOrFindUsersWithUAAPath(ccUsers []cf.User, path string) (updatedUsers []cf.User, apiResponse net.ApiResponse) {
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
		var ccUser cf.User

		for _, u := range ccUsers {
			if u.Guid == uaaResource.Id {
				ccUser = u
				break
			}
		}

		updatedUsers = append(updatedUsers, cf.User{
			Guid:     uaaResource.Id,
			Username: uaaResource.Username,
			IsAdmin:  ccUser.IsAdmin,
		})
	}
	return
}

func (repo CloudControllerUserRepository) Create(user cf.User) (apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
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
		user.Username,
		user.Username,
		user.Password,
		user.Username,
		user.Username,
	)
	request, apiResponse := repo.uaaGateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	type uaaUser struct {
		Id string
	}
	createUserResponse := &uaaUser{}

	_, apiResponse = repo.uaaGateway.PerformRequestForJSONResponse(request, createUserResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path = fmt.Sprintf("%s/v2/users", repo.config.Target)
	body = fmt.Sprintf(`{"guid":"%s"}`, createUserResponse.Id)
	return repo.ccGateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerUserRepository) Delete(user cf.User) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/users/%s", repo.config.Target, user.Guid)

	apiResponse = repo.ccGateway.DeleteResource(path, repo.config.AccessToken)
	if apiResponse.IsNotSuccessful() && apiResponse.ErrorCode != cf.USER_NOT_FOUND {
		return
	}

	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path = fmt.Sprintf("%s/Users/%s", uaaEndpoint, user.Guid)
	return repo.uaaGateway.DeleteResource(path, repo.config.AccessToken)
}

func (repo CloudControllerUserRepository) SetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse) {
	apiResponse = repo.setOrUnsetOrgRole("PUT", user, org, role)
	if apiResponse.IsNotSuccessful() {
		return
	}
	return repo.addOrgUserRole(user.Guid, org.Guid)
}

func (repo CloudControllerUserRepository) UnsetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse) {
	return repo.setOrUnsetOrgRole("DELETE", user, org, role)
}

func (repo CloudControllerUserRepository) setOrUnsetOrgRole(verb string, user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse) {
	rolePath, found := orgRoleToPathMap[role]

	if !found {
		apiResponse = net.NewApiResponseWithMessage("Invalid Role %s", role)
		return
	}

	path := fmt.Sprintf("%s/v2/organizations/%s/%s/%s", repo.config.Target, org.Guid, rolePath, user.Guid)

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

func (repo CloudControllerUserRepository) SetSpaceRole(user cf.User, space cf.Space, role string) (apiResponse net.ApiResponse) {
	rolePath, apiResponse := repo.checkSpaceRole(user, space, role)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.addOrgUserRole(user.Guid, space.Organization.Guid)
	if apiResponse.IsNotSuccessful() {
		return
	}

	return repo.ccGateway.UpdateResource(rolePath, repo.config.AccessToken, nil)
}

func (repo CloudControllerUserRepository) UnsetSpaceRole(user cf.User, space cf.Space, role string) (apiResponse net.ApiResponse) {
	rolePath, apiResponse := repo.checkSpaceRole(user, space, role)
	if apiResponse.IsNotSuccessful() {
		return
	}
	return repo.ccGateway.DeleteResource(rolePath, repo.config.AccessToken)
}

func (repo CloudControllerUserRepository) checkSpaceRole(user cf.User, space cf.Space, role string) (fullPath string, apiResponse net.ApiResponse) {
	rolePath, found := spaceRoleToPathMap[role]

	if !found {
		apiResponse = net.NewApiResponseWithMessage("Invalid Role %s", role)
	}

	fullPath = fmt.Sprintf("%s/v2/spaces/%s/%s/%s", repo.config.Target, space.Guid, rolePath, user.Guid)
	return
}

func (repo CloudControllerUserRepository) addOrgUserRole(userGuid, orgGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/organizations/%s/users/%s", repo.config.Target, orgGuid, userGuid)
	return repo.ccGateway.UpdateResource(path, repo.config.AccessToken, nil)
}
